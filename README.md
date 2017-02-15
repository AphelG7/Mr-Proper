# Mr-Proper

#### Requirements

* Telegram Bot (Privacy Mode disabled)
* MongoDB

#### Config

Keep a file called *mrproper.config* in the same directory as the executable.
It needs to look like this:

```
[Telegram]
	Token = "1234:token"
	Debug = false
	
[Mongo]
	Url = "mongodb://"
	Database = "mrproper"
	Debug = false
```
