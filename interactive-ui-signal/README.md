### Steps to run this sample:
1) You need a Temporal service running. See details in README.md
2) Run the following command to start the worker (registers both the proxy and the example workflow)
```
go run interactive-ui-signal/worker/main.go
```
3) Run the following command to start the web server
```
go run interactive-ui-signal/web/main.go
```
4) Open the web application at http://localhost:8080 and explore the sample:
    1. Create new accounts: starts a perpetual workflow for a new account
    2. Change the subscription type of the accounts: uses signal proxy workflow to communicate with the account workflow
    3. Delete account: uses signal proxy workflow to inform the account workflow to complete

#### Flow

```
     ┌──────┐                    ┌───┐                                ┌─────────────┐                                ┌───────────────┐
     │Client│                    │API│                                │ProxyWorkflow│                                │AccountWorkflow│
     └──┬───┘                    └─┬─┘                                └──────┬──────┘                                └───────┬───────┘
        │ POST /api/account/upgrade│                                         │                                               │        
        │ ─────────────────────────>                                         │                                               │        
        │                          │                                         │                                               │        
        │                          │ ExecuteWorkflow(signalType, signalInput)│                                               │        
        │                          │ ────────────────────────────────────────>                                               │        
        │                          │                                         │                                               │        
        │                          │                                         │ SignalExternalWorkflow(accountId, requestData)│        
        │                          │                                         │ ──────────────────────────────────────────────>        
        │                          │                                         │                                               │        
        │                          │                                         │ SignalExternalWorkflow(proxyId, responseData) │        
        │                          │                                         │ <──────────────────────────────────────────────        
        │                          │                                         │                                               │        
        │                          │         workflowExecution.Get()         │                                               │        
        │                          │ <────────────────────────────────────────                                               │        
        │                          │                                         │                                               │        
        │    write(responseData)   │                                         │                                               │        
        │ <─────────────────────────                                         │                                               │        
     ┌──┴───┐                    ┌─┴─┐                                ┌──────┴──────┐                                ┌───────┴───────┐
     │Client│                    │API│                                │ProxyWorkflow│                                │AccountWorkflow│
     └──────┘                    └───┘                                └─────────────┘                                └───────────────┘
```