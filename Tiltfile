
# Load the restart_process extension which enables in-place binary replacement
load('ext://restart_process', 'docker_build_with_restart')

### Secrets
k8s_yaml('./infra/development/k8s/secrets.yaml')

### Config
k8s_yaml('./infra/development/k8s/app-config.yaml')

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
    'domino/api-gateway',
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
    'domino/user-service',
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


### History Service
history_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/history-service ./services/history-service/cmd/main.go'
if os.name == 'nt':
    history_compile_cmd= './infra/development/docker/history-build.bat'

local_resource(
    'history-service-compile',
    history_compile_cmd,
    deps=['./services/history-service/', './shared'],
    labels="compiles"
)

docker_build_with_restart(
    'domino/history-service',
    '.',
    entrypoint=['./build/history-service'],
    dockerfile='./infra/development/docker/history-service.Dockerfile',
    only=[
       './build/history-service',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

k8s_yaml('./infra/development/k8s/history-service-deployment.yaml')
k8s_resource('history-service', resource_deps=['history-service-compile'], labels="services")


### lobby Service
lobby_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/lobby-service ./services/lobby-service/cmd/main.go'
if os.name == 'nt':
    lobby_compile_cmd= './infra/development/docker/lobby-build.bat'

local_resource(
    'lobby-service-compile',
    lobby_compile_cmd,
    deps=['./services/lobby-service/', './shared'],
    labels="compiles"
)

docker_build_with_restart(
    'domino/lobby-service',
    '.',
    entrypoint=['./build/lobby-service'],
    dockerfile='./infra/development/docker/lobby-service.Dockerfile',
    only=[
       './build/lobby-service',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

k8s_yaml('./infra/development/k8s/lobby-service-deployment.yaml')
k8s_resource('lobby-service', resource_deps=['lobby-service-compile'], labels="services")

### game Service
game_compile_cmd = 'CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/game-service ./services/game-service/cmd/main.go'
if os.name == 'nt':
    game_compile_cmd= './infra/development/docker/game-build.bat'

local_resource(
    'game-service-compile',
    game_compile_cmd,
    deps=['./services/game-service/', './shared'],
    labels="compiles"
)

docker_build_with_restart(
    'domino/game-service',
    '.',
    entrypoint=['./build/game-service'],
    dockerfile='./infra/development/docker/game-service.Dockerfile',
    only=[
       './build/game-service',
       './shared',
    ],
    live_update=[
        sync('./build', '/app/build'),
        sync('./shared', '/app/shared'),
    ],
)

k8s_yaml('./infra/development/k8s/game-service-deployment.yaml')
k8s_resource('game-service', resource_deps=['game-service-compile'], labels="services")


### Web Frontend ###
docker_build(
  'domino/web',
  '.',
  dockerfile='./infra/development/docker/web.Dockerfile',
)

k8s_yaml('./infra/development/k8s/web-deployment.yaml')
k8s_resource('web', port_forwards='3001:3000', labels="frontend")

### PostgreSQL DB 'domino'
