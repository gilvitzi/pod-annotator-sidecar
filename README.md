
[![GitHub license](https://img.shields.io/github/license/gilvitzi/pod-annotator-sidecar)](https://github.com/Naereen/StrapDown.js/blob/master/LICENSE)
[![Open in Visual Studio Code](https://open.vscode.dev/badges/open-in-vscode.svg)](https://open.vscode.dev/gilvitzi/pod-annotator-sidecar)



# Pod Annotator Sidecar Container for Kubernetes


Add this sidecar to your pod to allow your program to expose bussiness model data as annotations in the pod metadata.

Annotate your pods with useful information such as **progress**, **health** and **statistics** which can be later queried by cluster administrator or by an external controller or operator.

## How it works
Your app will create files in /annotations/{file-name}
where each file name is the annotation key and the file content is the annotation value (for annotation key="value").

for example, running the following command in your app container: `echo "my data" > /annotations/ data` will create a pod annotation: `data="my data"`

You can also configure all your annotations to have your app domain prefix (i.e `your.app.com/{filename}={file content}`)


![chart1](docs/images/pod-annotator-sidecar-chart.png?raw=true)
___
## Quick Start

1. clone this repo
2. run: 

`kubectl create -f deploy/`

3. check pods annotations using:

`kubectl get pod pod-annotator-example -o yaml`


## Usage example
* RBAC permissions are required, please see `role.yaml` and `rolebinding.yaml` in [deploy/](deploy/)

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: pod-annotator-example
spec:
  serviceAccountName: pod-annotator
  containers:
    - name: your-app
      image: k8s.gcr.io/busybox
      command: ["sh", "-c"]
      # sample app that will create and update 2 files in a loop
      args:
      - i=0; while true; do
          echo -n 'healthy' > /annotations/status;
          i=$((i+1));
          echo -n $i > /annotations/cycle;
          echo cycle $i;
          sleep 15;
        done;
      # mount the shared dir where the annotation files will be created
      volumeMounts:
        - name: annotations
          mountPath: /annotations
    - name: annotator-sidecar
      image: docker.io/gvitzi/pod-annotator
      command: ["pod-annotator"]
      # we set the --prefix so that each annotation will have this prefix in the begining of the key
      args:
        - "-v=5"
        - "--dir-path=/annotations"
        - "--prefix=my.app.com/"
      env:
        # inject the current namespace and pod name
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      # mount the shared dir where the annotation files are stored
      volumeMounts:
        - name: annotations
          mountPath: /annotations
  volumes:
    - name: annotations
      emptyDir: {}
```


## Build
If you want to make changes and build your own pod-annotator use the `make build` command to build the binary or `make docker-build` to build the docker image.
