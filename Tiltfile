
# Load the restart_process extension which enables in-place binary replacement
load('ext://restart_process', 'docker_build_with_restart')

### Secrets
k8s_yaml('./infra/development/k8s/secrets.yaml')

### RabbitMQ
k8s_yaml('./infra/development/k8s/rabbitmq-deployment.yaml')
k8s_resource('rabbitmq', port_forwards=['5672', '15672'], labels='tooling')

#### Jaeger
k8s_yaml('./infra/development/k8s/jaeger.yaml')
k8s_resource('jaeger', port_forwards=['16686:16686','14268:14268'], labels='tooling')

### API Gateway ###
# Compile
gateway_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/api-gateway ./services/api-gateway'
if os.name == 'nt':
    gateway_compile_cmd= './infra/development/docker/api-gateway-build.bat'
local_resource(
    'api-gateway-compile',
    gateway_compile_cmd,
    deps=['./services/api-gateway/', './shared/'],
    labels ="compiles"
)

# Build a docker image with enable live_update
docker_build_with_restart(
    'rebu/api-gateway',
    '.',
    entrypoint=['/app/build/api-gateway'],
    dockerfile='./infra/development/docker/api-gateway.Dockerfile',
    only=[
       './build/api-gateway',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

# Apply Kubernetes manifest
k8s_yaml('./infra/development/k8s/api-gateway-deployment.yaml')
k8s_resource('api-gateway', 
    port_forwards=8081,
    resource_deps=['api-gateway-compile', 'rabbitmq'],
    labels="services"
)

### Trip Service 
trip_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/trip-service ./services/trip-service/cmd/main.go'
if os.name == 'nt':
    trip_compile_cmd= './infra/development/docker/trip-build.bat'

local_resource(
    'trip-service-compile',
    trip_compile_cmd,
    deps=['./services/trip-service/', './shared'],
    labels="compiles"
)

docker_build_with_restart(
    'rebu/trip-service',
    '.',
    entrypoint=['./build/trip-service'],
    dockerfile='./infra/development/docker/trip-service.Dockerfile',
    only=[
       './build/trip-service',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

k8s_yaml('./infra/development/k8s/trip-service-deployment.yaml')
k8s_resource('trip-service', resource_deps=['trip-service-compile', 'rabbitmq'], labels="services")

### Driver Service 
driver_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/driver-service ./services/driver-service'
if os.name == 'nt':
    driver_compile_cmd= './infra/development/docker/driver-build.bat'

local_resource(
    'driver-service-compile',
    driver_compile_cmd,
    deps=['./services/driver-service/', './shared'],
    labels="compiles"
)

docker_build_with_restart(
    'rebu/driver-service',
    '.',
    entrypoint=['./build/driver-service'],
    dockerfile='./infra/development/docker/driver-service.Dockerfile',
    only=[
       './build/driver-service',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

k8s_yaml('./infra/development/k8s/driver-service-deployment.yaml')
k8s_resource('driver-service', resource_deps=['driver-service-compile', 'rabbitmq'], labels="services")

### User Service
user_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/user-service ./services/user-service/cmd/main.go'
if os.name == 'nt':
    user_compile_cmd= './infra/development/docker/user-build.bat'

local_resource(
    'user-service-compile',
    user_compile_cmd,
    deps=['./services/user-service/', './shared'],
    labels="compiles"
)

docker_build_with_restart(
    'rebu/user-service',
    '.',
    entrypoint=['./build/user-service'],
    dockerfile='./infra/development/docker/user-service.Dockerfile',
    only=[
       './build/user-service',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

k8s_yaml('./infra/development/k8s/user-service-deployment.yaml')
k8s_resource('user-service', resource_deps=['user-service-compile'], labels="services")

### Payment Service 
payment_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/payment-service ./services/payment-service/cmd/main.go'
if os.name == 'nt':
    payment_compile_cmd= './infra/development/docker/payment-build.bat'

local_resource(
    'payment-service-compile',
    payment_compile_cmd,
    deps=['./services/payment-service/', './shared'],
    labels="compiles"
)

docker_build_with_restart(
    'rebu/payment-service',
    '.',
    entrypoint=['./build/payment-service'],
    dockerfile='./infra/development/docker/payment-service.Dockerfile',
    only=[
       './build/payment-service',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

k8s_yaml('./infra/development/k8s/payment-service-deployment.yaml')
k8s_resource('payment-service', resource_deps=['payment-service-compile', 'rabbitmq'], labels="services")

### Web Frontend ###
docker_build(
  'rebu/web',
  '.',
  dockerfile='./infra/development/docker/web.Dockerfile',
)

k8s_yaml('./infra/development/k8s/web-deployment.yaml')
k8s_resource('web', port_forwards='3001:3000', labels="frontend")

### PostgreSQL DB 'domino'
