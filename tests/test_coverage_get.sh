#!/bin/bash

PATH_CURRUNT=$(pwd)
CODE_HOME="$PATH_CURRUNT/../"

function init_test_env()
{
    which gocov
    if [ $? -ne 0 ]; then
        go install github.com/axw/gocov/gocov@latest
    fi
}

function init_test_file()
{
    cd $CODE_HOME 2>&1 >/dev/null
    dirName=$(find . -maxdepth 10 -type d | grep -v '.git')
    for d in $dirName
    do
        cd $CODE_HOME/$d 2>&1 >/dev/null
        # 目录下没有go文件，不处理
        isHaveGoFile=$(ls -l | grep -E "*\.go$" | wc -l)
        if [ $isHaveGoFile -eq 0 ]; then
            cd $CODE_HOME 2>&1 >/dev/null
            continue
        fi
        # 目录下有test.go文件，不处理
        isHaveGoTestFile=$(ls -l | grep -E "*_test.go" | wc -l)
        if [ $isHaveGoTestFile -gt 0 ]; then
            cd $CODE_HOME 2>&1 >/dev/null
            continue
        fi
        # 生成test桩文件
        packageName=$(grep -h package $CODE_HOME/$d/*.go | head -n 1)
        echo "$packageName
import (
    \"testing\"
)

func TestMain(m *testing.M) {
    m.Run()
}" > go_keep_test.go
        cd $CODE_HOME 2>&1 >/dev/null
    done
}

function do_test()
{
    cd $CODE_HOME
    go test -gcflags=all=-l -coverpkg=./... -coverprofile=$CODE_HOME/cover.out ./...
    if [ $? -ne 0 ]; then
        echo "do unit test failed"
        exit 127
    fi

    gocov convert $CODE_HOME/cover.out > $CODE_HOME/gocov.json
    # 打印当前覆盖率统计信息
    gocov report $CODE_HOME/gocov.json
}

function main()
{
    init_test_env

    # 初始化代码目录，没有test用例的目录进行打桩
    init_test_file
    
    # 执行测试用例，并生成测试报告
    do_test
}

main
