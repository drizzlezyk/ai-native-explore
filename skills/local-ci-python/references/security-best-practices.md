# Python Security Best Practices

Common security issues detected by Bandit and how to fix them.

## Table of Contents

- [Hardcoded Credentials](#hardcoded-credentials)
- [SQL Injection](#sql-injection)
- [Shell Injection](#shell-injection)
- [Weak Cryptography](#weak-cryptography)
- [Insecure Deserialization](#insecure-deserialization)
- [Path Traversal](#path-traversal)
- [Debug Mode in Production](#debug-mode-in-production)
- [Insecure Random](#insecure-random)
- [Unvalidated Input](#unvalidated-input)

## Hardcoded Credentials

**Issue IDs**: B105, B106, B107

**Problem**: Hardcoded passwords, API keys, or secrets in source code.

**Bad**:
```python
# Hardcoded password
DB_PASSWORD = "secret123"
db = connect(password=DB_PASSWORD)

# Hardcoded API key
API_KEY = "sk-1234567890abcdef"
```

**Good**:
```python
import os

# Use environment variables
DB_PASSWORD = os.getenv('DB_PASSWORD')
if not DB_PASSWORD:
    raise ValueError("DB_PASSWORD environment variable not set")

db = connect(password=DB_PASSWORD)

# For API keys
API_KEY = os.getenv('API_KEY')
```

**Best Practice**:
- Store secrets in environment variables
- Use `.env` files (add to `.gitignore`)
- Use secret management services (AWS Secrets Manager, Azure Key Vault, HashiCorp Vault)
- Never commit secrets to version control

## SQL Injection

**Issue IDs**: B608, B609

**Problem**: User input concatenated directly into SQL queries.

**Bad**:
```python
# String formatting
query = f"SELECT * FROM users WHERE username = '{username}'"
cursor.execute(query)

# String concatenation
query = "SELECT * FROM users WHERE id = " + user_id
cursor.execute(query)
```

**Good**:
```python
# Parameterized queries (recommended)
query = "SELECT * FROM users WHERE username = %s"
cursor.execute(query, (username,))

# Or with named parameters
query = "SELECT * FROM users WHERE username = %(username)s"
cursor.execute(query, {'username': username})

# Using ORM (SQLAlchemy)
user = session.query(User).filter(User.username == username).first()
```

**Best Practice**:
- Always use parameterized queries
- Use ORM frameworks (SQLAlchemy, Django ORM)
- Never concatenate user input into SQL strings
- Validate and sanitize input as a secondary defense

## Shell Injection

**Issue IDs**: B602, B603, B604, B605, B606, B607

**Problem**: User input passed to shell commands without validation.

**Bad**:
```python
import subprocess
import os

# Using shell=True with user input
filename = request.form['filename']
subprocess.run(f'cat {filename}', shell=True)

# os.system with user input
os.system(f'rm {filename}')
```

**Good**:
```python
import subprocess
import shlex

# Use list form without shell=True
subprocess.run(['cat', filename])

# If you must use shell, validate input
allowed_files = ['file1.txt', 'file2.txt']
if filename in allowed_files:
    subprocess.run(['cat', filename])

# Or use shlex.quote() for proper escaping
safe_filename = shlex.quote(filename)
subprocess.run(f'cat {safe_filename}', shell=True)
```

**Best Practice**:
- Avoid `shell=True` whenever possible
- Use list form: `subprocess.run(['command', 'arg1', 'arg2'])`
- If shell=True is necessary, use `shlex.quote()` to escape arguments
- Validate input against allowlists
- Use Python libraries instead of shell commands when possible

## Weak Cryptography

**Issue IDs**: B303, B304, B305, B324, B501-B507

**Problem**: Using weak or insecure cryptographic algorithms.

**Bad**:
```python
import hashlib
import md5

# Weak hashing algorithms
hash = hashlib.md5(data).hexdigest()
hash = hashlib.sha1(data).hexdigest()

# Insecure random for security purposes
import random
token = random.randint(1000, 9999)
```

**Good**:
```python
import hashlib
import secrets

# Strong hashing algorithms
hash = hashlib.sha256(data).hexdigest()
hash = hashlib.sha512(data).hexdigest()

# For password hashing, use bcrypt or argon2
from bcrypt import hashpw, gensalt
password_hash = hashpw(password.encode('utf-8'), gensalt())

# Cryptographically secure random
token = secrets.token_hex(32)
verification_code = secrets.randbelow(10000)
```

**Best Practice**:
- Use SHA-256 or SHA-512 for hashing
- Use bcrypt, scrypt, or argon2 for password hashing
- Use `secrets` module for cryptographic randomness
- Never use MD5 or SHA-1 for security purposes
- Keep cryptographic libraries updated

## Insecure Deserialization

**Issue IDs**: B301, B302, B303, B304, B305, B306, B307, B308, B310

**Problem**: Deserializing untrusted data can lead to code execution.

**Bad**:
```python
import pickle
import yaml

# Pickle from untrusted source
data = pickle.loads(untrusted_data)

# YAML load (unsafe)
config = yaml.load(untrusted_yaml)
```

**Good**:
```python
import json
import yaml

# Use JSON for data serialization
data = json.loads(untrusted_data)

# Use safe YAML loading
config = yaml.safe_load(untrusted_yaml)

# If you must use pickle, verify source
# And consider signing/encrypting the data
import hmac
import hashlib

def safe_pickle_load(data, secret_key):
    signature, pickled = data[:32], data[32:]
    expected = hmac.new(secret_key, pickled, hashlib.sha256).digest()
    if not hmac.compare_digest(signature, expected):
        raise ValueError("Invalid signature")
    return pickle.loads(pickled)
```

**Best Practice**:
- Use JSON for data serialization (safe by design)
- Use `yaml.safe_load()` instead of `yaml.load()`
- Avoid pickle for untrusted data
- If pickle is necessary, verify data integrity with HMAC
- Consider using safer alternatives like msgpack or protobuf

## Path Traversal

**Issue IDs**: B108, B113

**Problem**: User-controlled file paths can access unauthorized files.

**Bad**:
```python
# User input directly in file path
filename = request.args.get('file')
with open(f'/var/www/uploads/{filename}', 'r') as f:
    content = f.read()

# Can access: /var/www/uploads/../../../etc/passwd
```

**Good**:
```python
import os
from pathlib import Path

# Validate and sanitize file paths
def safe_file_read(filename):
    # Remove any path components
    filename = os.path.basename(filename)

    # Construct safe path
    base_dir = Path('/var/www/uploads')
    file_path = base_dir / filename

    # Resolve and verify path is within base_dir
    try:
        file_path = file_path.resolve()
        file_path.relative_to(base_dir)
    except (ValueError, RuntimeError):
        raise ValueError("Invalid file path")

    with open(file_path, 'r') as f:
        return f.read()

# Or use allowlist
ALLOWED_FILES = {'report.pdf', 'data.csv'}
if filename in ALLOWED_FILES:
    with open(f'/var/www/uploads/{filename}', 'r') as f:
        content = f.read()
```

**Best Practice**:
- Use `os.path.basename()` to strip directory components
- Validate resolved paths stay within allowed directories
- Use allowlists for file names when possible
- Avoid user input in file paths entirely if possible

## Debug Mode in Production

**Issue IDs**: B201, B701

**Problem**: Running Flask/Django with debug mode in production exposes sensitive information.

**Bad**:
```python
from flask import Flask

app = Flask(__name__)
app.run(debug=True)  # Debug mode enabled

# Django settings.py
DEBUG = True
```

**Good**:
```python
import os
from flask import Flask

app = Flask(__name__)

# Only debug in development
if os.getenv('FLASK_ENV') == 'development':
    app.run(debug=True)
else:
    app.run(debug=False)

# Django settings.py
DEBUG = os.getenv('DJANGO_DEBUG', 'False') == 'True'
```

**Best Practice**:
- Never enable debug mode in production
- Use environment variables to control debug settings
- Implement proper logging instead of relying on debug output
- Use error tracking services (Sentry, Rollbar) for production errors

## Insecure Random

**Issue ID**: B311

**Problem**: Using `random` module for security-sensitive operations.

**Bad**:
```python
import random

# Security token (predictable!)
token = random.randint(100000, 999999)

# Session ID (predictable!)
session_id = ''.join(random.choices('0123456789abcdef', k=32))
```

**Good**:
```python
import secrets

# Cryptographically secure token
token = secrets.randbelow(900000) + 100000

# Cryptographically secure session ID
session_id = secrets.token_hex(32)

# Or use token_urlsafe for URL-safe tokens
api_key = secrets.token_urlsafe(32)
```

**Best Practice**:
- Use `secrets` module for any security-sensitive randomness
- Never use `random` for tokens, passwords, or cryptographic purposes
- `secrets` is cryptographically strong and suitable for security

## Unvalidated Input

**Issue IDs**: B110, B112, B113

**Problem**: Using user input without validation.

**Bad**:
```python
# Direct use of user input
redirect_url = request.args.get('next')
return redirect(redirect_url)

# User-controlled regex
pattern = request.form['pattern']
re.compile(pattern)  # ReDoS vulnerability
```

**Good**:
```python
from urllib.parse import urlparse

# Validate redirect URLs
def safe_redirect(url):
    parsed = urlparse(url)
    # Only allow relative URLs or same domain
    if parsed.netloc and parsed.netloc != request.host:
        raise ValueError("Invalid redirect URL")
    return redirect(url)

# Validate and limit regex complexity
MAX_PATTERN_LENGTH = 100
if len(pattern) > MAX_PATTERN_LENGTH:
    raise ValueError("Pattern too long")

# Use timeout for regex
import regex  # pip install regex
compiled = regex.compile(pattern, timeout=1.0)
```

**Best Practice**:
- Validate all user input
- Use allowlists instead of blocklists
- Sanitize input before use
- Implement rate limiting for user-controlled operations
- Set timeouts for potentially expensive operations

## Additional Resources

- [OWASP Python Security](https://cheatsheetseries.owasp.org/cheatsheets/Python_Security_Cheat_Sheet.html)
- [Bandit Documentation](https://bandit.readthedocs.io/)
- [PyCQA Security Best Practices](https://github.com/PyCQA/bandit)
- [NIST Secure Coding Guidelines](https://www.nist.gov/publications/secure-coding-guidelines)
