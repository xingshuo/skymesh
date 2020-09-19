
#为了消除不同protoc版本生成的pb.go不同引入的问题, 所有proto生成的pb.go统一以生成好的源码方式提供, 禁止修改
#echo "start building proto..."
#protoc --proto_path=./proto/ --go_out=./proto/generate ./proto/*.proto
#protoc --proto_path=./examples/inner_service --go_out=./examples/inner_service ./examples/inner_service/*.proto
#protoc --proto_path=./examples/grpc/helloworld --go_out=./examples/grpc/helloworld ./examples/grpc/helloworld/*.proto
#echo "build proto done."

echo "start building..."
cd nameserver/bootstrap
rm nameserver.exe
go build -gcflags "-N -l" -mod=vendor -o nameserver.exe main.go
cd ../../
echo "build nameserver done."

cd examples/helloworld/greeter_client
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build greeter client done."

cd examples/helloworld/greeter_server
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build greeter server done."

cd examples/features/name_router/client
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../../
echo "build features name_router client done"

cd examples/features/name_router/server
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../../
echo "build features name_router server done"

cd examples/inner_service/client
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build inner_service client done"

cd examples/inner_service/server
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../
echo "build inner_service server done"

cd examples/grpc/helloworld/greeter_client
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../../
echo "build grpc greeter client done"

cd examples/grpc/helloworld/greeter_server
rm main.exe
go build -gcflags "-N -l" -mod=vendor -o main.exe main.go
cd ../../../../
echo "build grpc greeter server done"