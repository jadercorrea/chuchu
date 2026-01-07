
# JavaScript Patterns

Guidelines for writing modern, maintainable JavaScript code.

## When to Activate

- Writing or editing JavaScript code (`.js`, `.mjs` files)
- Creating Node.js applications
- Reviewing JavaScript code

## Modern Syntax

### Use const/let, never var

```javascript
// GOOD
const MAX_RETRIES = 3;
let attempts = 0;

// BAD
var MAX_RETRIES = 3;
var attempts = 0;
```

### Use arrow functions for callbacks

```javascript
// GOOD
const squares = numbers.map(n => n ** 2);
const adults = users.filter(u => u.age >= 18);

// BAD
const squares = numbers.map(function(n) {
  return n ** 2;
});
```

### Use destructuring

```javascript
// GOOD - object destructuring
const { name, email, role = 'user' } = user;

// GOOD - array destructuring
const [first, second, ...rest] = items;

// GOOD - parameter destructuring
function createUser({ name, email, admin = false }) {
  return { name, email, admin };
}
```

### Use template literals

```javascript
// GOOD
const message = `Hello, ${user.name}! You have ${count} notifications.`;
const html = `
  <div class="card">
    <h2>${title}</h2>
    <p>${description}</p>
  </div>
`;

// BAD
const message = 'Hello, ' + user.name + '! You have ' + count + ' notifications.';
```

### Use spread operator

```javascript
// GOOD - arrays
const combined = [...arr1, ...arr2];
const copy = [...original];

// GOOD - objects
const updated = { ...user, name: 'New Name' };
const merged = { ...defaults, ...options };
```

## Functions

### Use default parameters

```javascript
// GOOD
function createUser(name, role = 'user', active = true) {
  return { name, role, active };
}

// BAD
function createUser(name, role, active) {
  role = role || 'user';
  active = active !== undefined ? active : true;
  return { name, role, active };
}
```

### Use rest parameters

```javascript
// GOOD
function sum(...numbers) {
  return numbers.reduce((a, b) => a + b, 0);
}

// BAD
function sum() {
  return Array.from(arguments).reduce((a, b) => a + b, 0);
}
```

## Async/Await

### Prefer async/await over promises

```javascript
// GOOD
async function fetchUserData(userId) {
  try {
    const user = await fetchUser(userId);
    const orders = await fetchOrders(user.id);
    return { user, orders };
  } catch (error) {
    console.error('Failed to fetch:', error);
    throw error;
  }
}

// BAD - promise chains
function fetchUserData(userId) {
  return fetchUser(userId)
    .then(user => fetchOrders(user.id)
      .then(orders => ({ user, orders })))
    .catch(error => {
      console.error('Failed to fetch:', error);
      throw error;
    });
}
```

### Use Promise.all for parallel operations

```javascript
// GOOD - parallel
const [users, products, orders] = await Promise.all([
  fetchUsers(),
  fetchProducts(),
  fetchOrders(),
]);

// BAD - sequential when not needed
const users = await fetchUsers();
const products = await fetchProducts();
const orders = await fetchOrders();
```

## Error Handling

### Use custom errors

```javascript
class NotFoundError extends Error {
  constructor(resource, id) {
    super(`${resource} with id ${id} not found`);
    this.name = 'NotFoundError';
    this.resource = resource;
    this.id = id;
  }
}

class ValidationError extends Error {
  constructor(field, message) {
    super(`${field}: ${message}`);
    this.name = 'ValidationError';
    this.field = field;
  }
}
```

### Always catch async errors

```javascript
// GOOD
async function handleRequest(req, res) {
  try {
    const data = await processRequest(req);
    res.json(data);
  } catch (error) {
    if (error instanceof ValidationError) {
      res.status(400).json({ error: error.message });
    } else {
      res.status(500).json({ error: 'Internal server error' });
    }
  }
}
```

## Arrays

### Use array methods over loops

```javascript
// GOOD
const active = users.filter(u => u.active);
const names = users.map(u => u.name);
const total = items.reduce((sum, item) => sum + item.price, 0);
const hasAdmin = users.some(u => u.role === 'admin');
const allActive = users.every(u => u.active);

// BAD
const active = [];
for (let i = 0; i < users.length; i++) {
  if (users[i].active) {
    active.push(users[i]);
  }
}
```

### Use find for single items

```javascript
// GOOD
const admin = users.find(u => u.role === 'admin');

// BAD
const admin = users.filter(u => u.role === 'admin')[0];
```

## Objects

### Use shorthand properties

```javascript
// GOOD
const name = 'Alice';
const age = 30;
const user = { name, age };

// BAD
const user = { name: name, age: age };
```

### Use computed property names

```javascript
// GOOD
const key = 'dynamicKey';
const obj = {
  [key]: 'value',
  [`prefix_${key}`]: 'other value',
};
```

### Use Object.entries/values/keys

```javascript
// Iterate over object
for (const [key, value] of Object.entries(obj)) {
  console.log(`${key}: ${value}`);
}

// Transform object to array
const values = Object.values(config);
const keys = Object.keys(config);
```

## Modules

### Use ES modules

```javascript
// GOOD - named exports
export function formatDate(date) { ... }
export const MAX_SIZE = 1024;

// GOOD - default export for main item
export default class UserService { ... }

// GOOD - imports
import UserService, { formatDate, MAX_SIZE } from './user-service.js';
```

### Organize exports in index files

```javascript
// utils/index.js
export { formatDate, parseDate } from './date.js';
export { formatCurrency } from './currency.js';
export { validateEmail, validatePhone } from './validation.js';

// Usage
import { formatDate, formatCurrency, validateEmail } from './utils/index.js';
```

## Optional Chaining and Nullish Coalescing

```javascript
// GOOD - optional chaining
const city = user?.address?.city;
const first = items?.[0];
const result = obj?.method?.();

// GOOD - nullish coalescing
const name = user.name ?? 'Anonymous';
const count = data.count ?? 0;

// BAD
const city = user && user.address && user.address.city;
const name = user.name || 'Anonymous'; // Wrong for empty string
```
