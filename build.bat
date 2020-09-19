@echo "start building..."

cd nameserver\bootstrap
del nameserver.exe
go build -gcflags "-N -l" -mod=vendor -o nameserver.exe main.go
cd ..\..\
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

cd examples\features\name_router\client
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\..\
@echo "build features name_router client done"

cd examples\features\name_router\server
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\..\
@echo "build features name_router server done"

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

cd examples\features\election\client
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\..\
@echo "build features election client done"

cd examples\features\election\server
del main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ..\..\..\..\
@echo "build features election server done"