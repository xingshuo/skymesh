@echo "start building..."

cd nameserver\bootstrap
del nameserver.exe
go build -gcflags "-N -l" -mod=vendor -o nameserver.exe main.go
cd ..\..\
copy nameserver\bootstrap\nameserver.exe examples\helloworld\nameserver\main.exe
copy nameserver\bootstrap\nameserver.exe examples\nameservice\nameserver\main.exe
copy nameserver\bootstrap\nameserver.exe examples\inner_service\nameserver\main.exe
copy nameserver\bootstrap\nameserver.exe examples\grpc\helloworld\nameserver\main.exe
@echo "build nameserver done."

cd examples\helloworld\greeter_client
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\
@echo "build greeter client done."

cd examples\helloworld\greeter_server
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\
@echo "build greeter server done."

cd examples\nameservice\client
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\
@echo "build nameservice client done"

cd examples\nameservice\server
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\
@echo "build nameservice server done"

cd examples\inner_service\client
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\
@echo "build inner_service client done"

cd examples\inner_service\server
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\
@echo "build inner_service server done"

cd examples\grpc\helloworld\greeter_client
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\..\
@echo "build grpc greeter client done"

cd examples\grpc\helloworld\greeter_server
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\..\
@echo "build grpc greeter server done"