# -*- mode: Python -*-

load('ext://uibutton', 'cmd_button')
load('ext://namespace', 'namespace_create')

cmd_button(name='disable all',
        argv=['tilt', 'disable', '--all'],
        text='disable all',
        location='nav',
        icon_name='swipe_down',
)

cmd_button(name='enable all',
        argv=['tilt', 'enable', '--all'],
        text='enable all',
        location='nav',
        icon_name='swipe_up',
)

namespace_create('datree')
DATREE_TOKEN=os.environ.get('DATREE_TOKEN')
k8s_yaml(helm('./charts/datree-admission-webhook/', name='admission-webhook', values='internal/fixtures/values.dev.yaml', namespace='datree', set=['devMode.enabled=true ','securityContext.runAsUser=1000', 'securityContext.readOnlyRootFilesystem=false', 'clusterScanner.image.pullPolicy=Never', "datree.token={}".format(DATREE_TOKEN)]))



def debugging():
    docker_build('webhook-server', './', dockerfile = './Dockerfile.debugging.tilt', build_args={
        "BUILD_ENVIRONMENT":"staging",
        "WEBHOOK_VERSION":"0.0.1",
    })

    docker_build('datree/cluster-scanner-staging', '../cluster-scanner', dockerfile = '../cluster-scanner/Dockerfile.debugging.tilt', build_args={
        "BUILD_ENVIRONMENT":"staging",
        "SCANNER_VERSION":"0.0.1",
    })

    local_resource(
    name='datree-webhook-server-debuging',
    serve_cmd='bash ./tilt/scripts/port-forwarding-for-debugging.sh "datree-webhook-server" "5555"',
    )

    local_resource(
        name='cluster-scanner debugging',
        serve_cmd='bash ./tilt/scripts/port-forwarding-for-debugging.sh "datree-cluster-scanner-server" "5556"',
    )

    local_resource(
        name='webhook - restart port-forwarding',
        serve_cmd='bash ./tilt/scripts/restart-port-forwarding.sh "datree-webhook-server" "datree-webhook-server-debuging"',
    )

    local_resource(
        name='cluster-scanner - restart port-forwarding',
        serve_cmd='bash ./tilt/scripts/restart-port-forwarding.sh "datree-cluster-scanner-server" "cluster-scanner debugging"',
    )

def hot_reload():
    docker_build('webhook-server', './', dockerfile = './Dockerfile',
        build_args={
            "BUILD_ENVIRONMENT":"staging",
            "WEBHOOK_VERSION":"0.0.1",
        }
    )


    docker_build('datree/cluster-scanner-staging', '../cluster-scanner', dockerfile = '../cluster-scanner/Dockerfile',
        build_args={
            "BUILD_ENVIRONMENT":"staging",
            "SCANNER_VERSION":"0.0.1",
        }
    )
    

if os.environ.get('TILT_ENV') == 'debugging':
    debugging()
    
else:
   hot_reload()
