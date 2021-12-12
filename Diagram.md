 ```
 User request                               User request
     |                                         |
     |            example.service.v1           |       example.service.v2        example.service.private
+----+------------------------------+     +----+------------------------+   +---------------------------+
|    v                              |     |    v                        |   |                           |
|  Handler                          |     |  Handler         +----------+---+-->Handler                 |
|     |                             |     |     |            |          |   |      |                    |
|     v                             |     |     v            |          |   |      v                    |
|  Input validation                 |     |  Input validation|          |   |   Input validation        |
|     |                             |     |     |            |          |   |      |                    |
|     v                             |     |     v            |          |   |      v                    |
|  Impl method-------+            +-+-+---+->Impl method     |          |   |   Service implementation  |
|   |                |            | | |   |     |            |          |   |                           |
|   v                v            | | |   |     v            |          |   +---------------------------+
|  Input converter  Append mutators | |   |  Input converter |          |
|               |                   | |   |     |            |          |
|               +-------------------+-+   |     v            |          |
|                                   |     |  Apply mutators--+          |
|                                   |     |                             |
+-----------------------------------+     +-----------------------------+
```
