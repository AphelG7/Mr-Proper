# Mr. Proper

If you want to see an instance of the bot in action have a look at [@ProperlyGroupBot](https://t.me/ProperlyGroupBot).

### Requirements

* Telegram Bot (Privacy Mode disabled)
* MongoDB

### Executables

You can find the latest binaries for the bot [here](https://github.com/4m4rOk/Mr-Proper/releases).

### Configuration

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
