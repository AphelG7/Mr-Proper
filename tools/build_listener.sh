./fmt_proper.sh
echo "Building listener for darwin i386..."
GOOS=darwin GOARCH=386 go build -o ../../../../../bin/Mr-Proper/MrProper_darwin_386 github.com/4m4rOk/Mr-Proper/listener
echo "Building listener for linux arm..."
GOOS=linux GOARCH=arm go build -o ../../../../../bin/Mr-Proper/MrProper_linux_arm github.com/4m4rOk/Mr-Proper/listener
echo "Building listener for linux i386..."
GOOS=linux GOARCH=386 go build -o ../../../../../bin/Mr-Proper/MrProper_linux_386 github.com/4m4rOk/Mr-Proper/listener
echo "Building listener for linux amd64..."
GOOS=linux GOARCH=amd64 go build -o ../../../../../bin/Mr-Proper/MrProper_linux_amd64 github.com/4m4rOk/Mr-Proper/listener
echo "Building listener for windows i386..."
GOOS=windows GOARCH=386 go build -o ../../../../../bin/Mr-Proper/MrProper_linux_win32 github.com/4m4rOk/Mr-Proper/listener
echo "Building listener for windows amd64..."
GOOS=windows GOARCH=amd64 go build -o ../../../../../bin/Mr-Proper/MrProper_win64 github.com/4m4rOk/Mr-Proper/listener