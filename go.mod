module github.com/omatztw/gomatalk

go 1.12

require (
	github.com/boltdb/bolt v1.3.1
	github.com/bwmarrin/discordgo v0.28.1
	github.com/fsnotify/fsnotify v1.6.0
	github.com/golang-migrate/migrate/v4 v4.15.2
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/omatztw/dgvoice v0.0.0-20220223044428-1f960e96950b
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.7 // indirect
	github.com/spf13/afero v1.9.5 // indirect
	github.com/spf13/viper v1.15.0
	go.uber.org/atomic v1.10.0 // indirect
	golang.org/x/crypto v0.7.0 // indirect
	gorm.io/driver/sqlite v1.4.4
	gorm.io/gorm v1.24.6
	layeh.com/gopus v0.0.0-20210501142526-1ee02d434e32 // indirect
)

replace github.com/bwmarrin/discordgo v0.28.1 => github.com/omatztw/discordgo v0.0.0-20251126152932-a505ba4d1fd4
