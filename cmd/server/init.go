package main

import (
	_ "server/internal/service/comm/db"
	_ "server/internal/service/cross/cross"
	_ "server/internal/service/game/actor"
	_ "server/internal/service/game/game"
	_ "server/internal/service/gate"
	_ "server/internal/service/platform/access"
	_ "server/internal/service/platform/account"
	_ "server/internal/service/platform/auth"
	_ "server/internal/service/realm/scene"
	_ "server/internal/service/realm/scenemgr"
	_ "server/internal/service/social/social"
)
