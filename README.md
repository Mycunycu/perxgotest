# perxgotest

#### The service for calculating sum all elements arithmetic progression.

**Tasks are executed from the queue.\
Lifecycle a task: task -> queue -> processing -> tasks history.
At the same time can be processing only a pointed number of tasks.**

### How to use?

**Just run the application with flag `N` - it's a maximum number parallel task processing at the one time (by default - one process).**
**The Service will start on the port `:8080`**

### Endpoints:

- **Enqueue a task**
  **POST** `/api/v1/task`
  Body: `{ "element_amount": 10, "delta": -2.784, "first_element": 12.4589, "interval": 0.547, "ttl": 500000 }`
  Where:
  `element_amount` - count of elements in progression(integer);
  `delta` - step value between elements(float);
  `first_element` - the value of first element(float);
  `interval` - delay between iterations in sec(float);
  `ttl` - time to live in history in sec(float);

- **Get sorted tasks list**
  **GET** `/api/v1/task`
  The list sorted in a first by queue position and second by enqueue time for task if already done.
