@echo "start building..."
protoc --proto_path=.\proto\ --go_out=.\proto\generate .\proto\*.proto
@echo "build proto done."

cd nameserver\bootstrap
del nameserver.exe
go build -gcflags "-N -l" -o nameserver.exe main.go
cd ..\..\
copy nameserver\bootstrap\nameserver.exe examples\helloworld\nameserver\main.exe
copy nameserver\bootstrap\nameserver.exe examples\nameservice\nameserver\main.exe
@echo "build nameserver done."

cd examples\helloworld\greeter_client
del main.exe
go build -gcflags "-N -l" -o main.exe main.go
cd ..\..\..\
@echo "build greeter client done."

cd examples\helloworld\greeter_server
del main.exe
go build -gcflags "-N -l" -o main.exe main.go
cd ..\..\..\
@echo "build greeter server done."

cd examples\nameservice\client
del main.exe
go build -gcflags "-N -l" -o main.exe main.go
cd ..\..\..\
@echo "build nameservice client done"

cd examples\nameservice\server
del main.exe
go build -gcflags "-N -l" -o main.exe main.go
cd ..\..\..\
@echo "build nameservice server done"