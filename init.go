package main

import _ "server/service/comm/db"
import _ "server/service/gate"

//import _ "server/service/game/game"
import _ "server/service/world/zone"

// 服务注册：新增服务时，添加 blank import 触发 init() 自动注册
