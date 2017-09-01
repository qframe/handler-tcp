# handler-tcp
Handler to send messages via TCP

```bash
$ go run main.go                                                                                                                                                                                                                                        git:(master|✚5
2017/09/01 19:32:25.929263 [  INFO] Dispatch broadcast for Data, Done and Tick
2017/09/01 19:32:25.929431 [NOTICE]             tcp Name:loop       >> Start collector v0.2.4
2017/09/01 19:32:25.929442 [NOTICE]         tcp-out Name:out        >> Start log handler v0.1.5
2017/09/01 19:32:25.929532 [NOTICE]             tcp Name:in         >> Start collector v0.2.4
2017/09/01 19:32:25.929594 [NOTICE]             log Name:log        >> Start log handler v0.2.0
2017/09/01 19:32:25.929715 [  INFO]             tcp Name:loop       >> Listening on 0.0.0.0:10002
2017/09/01 19:32:25.929763 [  INFO]             tcp Name:in         >> Listening on 0.0.0.0:10001
2017/09/01 19:32:25.929954 [  INFO]         tcp-out Name:out        >> Connected to '127.0.0.1:10002'
```

When sending to the inbound TCP socket via `export CNT=$((${CNT:-0}+1)) ; echo "Test${CNT}"|nc localhost 10001` the messages is routed through the agent.

 * Received by `in` (collector-tcp)
 * `in` is processed by `out` (handler-tcp)
 * `out` is received by `loop` (collector-tcp)
 * and the message received by `loop` is printed by the `log` (handler-log)

As the message is not accociated with a container IP, the inventory-request times out.

```bash
2017/09/01 19:32:30.031812 [ DEBUG]             tcp Name:in         >> Experience timeout for IP localhost... continue w/o Container info (SourcePath: in)
2017/09/01 19:32:30.031934 [ DEBUG]         tcp-out Name:out        >> Sending 'Test1'
2017/09/01 19:32:30.032069 [ DEBUG]             log Name:log        >> InputsMatch([loop]) != in
2017/09/01 19:32:32.035469 [ DEBUG]             tcp Name:loop       >> Experience timeout for IP 127.0.0.1... continue w/o Container info (SourcePath: loop)
2017/09/01 19:32:32.035573 [  INFO]             log Name:log        >> loop            : Test1
2017/09/01 19:32:32.035608 [ DEBUG]         tcp-out Name:out        >> InputsMatch([in]) != loop
```
