# Redis Queue

Redis queue is the wrapper of github.com/hibiken/asynq package.

## Notable Differences

To follow the common queue implementation, we wrap the package `asynq` with the following differences:

- In `asynq`, one queue can handle different tasks. In this library, we have a separate queue for each task (basically, term "task" is not used).
- In `asynq`, one subscriber can listen to multiple queues & tasks. In this library, we have a separate subscriber for each queue.

## Example Usage

Check the example in the [examples/redisqueue](../examples/redisqueue) directory.
