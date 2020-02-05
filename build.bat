@echo "start building..."
protoc --proto_path=.\proto\ --go_out=.\proto\generate .\proto\*.proto
@echo "build proto done."
cd nameserver\bootstrap
del nameserver.exe
go build -o nameserver.exe main.go
cd ..\..\
@echo "build nameserver done."