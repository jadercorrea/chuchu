---
layout: skill
title: Python Idioms
name: python
language: python
description: Write idiomatic Python code following PEP standards and community best practices.
---

# Python Idioms

Guidelines for writing idiomatic, maintainable Python code.

## When to Activate

- Writing or editing Python code (`.py` files)
- Creating new Python modules or packages
- Reviewing Python code

## General Style

### Follow PEP 8

```python
# GOOD - PEP 8 compliant
def calculate_total(items, tax_rate=0.08):
    """Calculate total price with tax."""
    subtotal = sum(item.price for item in items)
    return subtotal * (1 + tax_rate)

# BAD - violates PEP 8
def CalculateTotal(Items,TaxRate=0.08):
    SubTotal=sum(Item.price for Item in Items)
    return SubTotal*(1+TaxRate)
```

### Use f-strings for formatting

```python
# GOOD - f-strings (Python 3.6+)
message = f"Hello, {user.name}! You have {count} notifications."

# BAD - old-style formatting
message = "Hello, %s! You have %d notifications." % (user.name, count)
message = "Hello, {}! You have {} notifications.".format(user.name, count)
```

## Data Structures

### Use list/dict/set comprehensions

```python
# GOOD
squares = [x**2 for x in range(10)]
user_map = {u.id: u for u in users}
unique_names = {u.name.lower() for u in users}

# BAD
squares = []
for x in range(10):
    squares.append(x**2)
```

### Use unpacking

```python
# GOOD
first, *middle, last = items
a, b = b, a  # swap
name, age, email = user_tuple

# BAD
first = items[0]
middle = items[1:-1]
last = items[-1]
```

### Use defaultdict and Counter

```python
from collections import defaultdict, Counter

# GOOD
word_count = Counter(words)
groups = defaultdict(list)
for item in items:
    groups[item.category].append(item)

# BAD
word_count = {}
for word in words:
    word_count[word] = word_count.get(word, 0) + 1
```

## Functions

### Use keyword arguments for clarity

```python
# GOOD
def send_email(to, subject, body, *, cc=None, bcc=None, reply_to=None):
    ...

send_email(
    to="user@example.com",
    subject="Hello",
    body="Content",
    reply_to="noreply@example.com"
)

# BAD
send_email("user@example.com", "Hello", "Content", None, None, "noreply@example.com")
```

### Use type hints

```python
from typing import Optional, List

def find_user(user_id: int) -> Optional[User]:
    """Find a user by ID."""
    return db.query(User).filter(User.id == user_id).first()

def process_orders(orders: List[Order]) -> dict[str, int]:
    """Process orders and return summary."""
    return {"total": len(orders), "completed": sum(1 for o in orders if o.done)}
```

### Use generators for large data

```python
# GOOD - memory efficient
def read_large_file(path):
    with open(path) as f:
        for line in f:
            yield line.strip()

# BAD - loads entire file
def read_large_file(path):
    with open(path) as f:
        return f.readlines()
```

## Error Handling

### Be specific with exceptions

```python
# GOOD
try:
    user = users[user_id]
except KeyError:
    raise UserNotFoundError(f"User {user_id} not found")

# BAD
try:
    user = users[user_id]
except:
    raise Exception("Error")
```

### Use context managers

```python
# GOOD
with open("data.json") as f:
    data = json.load(f)

with db.transaction():
    user.save()
    order.save()

# BAD
f = open("data.json")
data = json.load(f)
f.close()
```

## Classes

### Use dataclasses for data containers

```python
from dataclasses import dataclass
from typing import Optional

# GOOD
@dataclass
class User:
    id: int
    name: str
    email: str
    admin: bool = False

# BAD - boilerplate
class User:
    def __init__(self, id, name, email, admin=False):
        self.id = id
        self.name = name
        self.email = email
        self.admin = admin
```

### Use properties for computed attributes

```python
class Order:
    def __init__(self, items):
        self.items = items
    
    @property
    def total(self):
        return sum(item.price for item in self.items)
    
    @property
    def is_empty(self):
        return len(self.items) == 0
```

## Testing

### Use pytest fixtures

```python
import pytest

@pytest.fixture
def user():
    return User(id=1, name="Test", email="test@example.com")

@pytest.fixture
def db_session():
    session = create_session()
    yield session
    session.rollback()

def test_user_creation(db_session, user):
    db_session.add(user)
    assert db_session.query(User).count() == 1
```

### Use parametrized tests

```python
@pytest.mark.parametrize("input,expected", [
    ("hello", "HELLO"),
    ("World", "WORLD"),
    ("", ""),
])
def test_uppercase(input, expected):
    assert input.upper() == expected
```

## Pythonic Patterns

### Use enumerate for index + value

```python
# GOOD
for i, item in enumerate(items):
    print(f"{i}: {item}")

# BAD
for i in range(len(items)):
    print(f"{i}: {items[i]}")
```

### Use zip for parallel iteration

```python
# GOOD
for name, score in zip(names, scores):
    print(f"{name}: {score}")

# BAD
for i in range(len(names)):
    print(f"{names[i]}: {scores[i]}")
```

### Use any/all for conditions

```python
# GOOD
if any(item.expired for item in items):
    raise ValueError("Some items are expired")

if all(user.verified for user in users):
    send_notification()

# BAD
has_expired = False
for item in items:
    if item.expired:
        has_expired = True
        break
```
