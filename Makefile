build_all: build_scale_tools_jar build_printer_tools_jar build_linux_amd64 build_windows_amd64 build_darwin_amd64 build_darwin_arm64

build_scale_tools_jar:
	scale-tools/gradlew -p ./scale-tools tasks jar

build_printer_tools_jar:
	printer-tools/gradlew -p ./printer-tools tasks jar

build_linux_amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/companion_linux_amd64

build_windows_amd64:
	GOOS=windows GOARCH=amd64 go build -o bin/companion_windows_amd64.exe

build_darwin_amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/companion_darwin_amd64

build_darwin_arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/companion_darwin_arm64

checksums:
	cd ./bin && sha512sum * > checksums.txt
