# NanoPod

NanoPod help you to patch your pods automatically without changing your Deployment/StatefulSet/Job/ReplicaSet etc.

## Description

The NanoPods for pods is like the nano-armor for the Iron-Man, It can automatically arm your pods without changing Deployment/StatefulSet/Job/ReplicaSet etc. Only the pods labeled with "nano-pods" would be accessed to. It provides an aspect oriented configuration.

## Getting Started

NanoPod Operator required [cert-manager](https://cert-manager.io/docs/installation/), So firstly make sure you have [cert-manager](https://cert-manager.io/docs/installation/) installed

Then run the following scripts in your Kubernetes, to install NanoPod Operator and related CRDs, register webhooks.
```shell
kubectl apply -f https://github.com/Duncan-tree-zhou/nano-pod/releases/download/v0.0.1/nano-pod-operator.yaml
```

After the `nano-pod-controller-manager` deployment is ready, the functionality of NanoPod is enabled. 

Then create namespace `nano-pod-test` for testing, and apply the following yaml to create a NanoPod: 

```yaml
# my-nano-pod.yaml
apiVersion: nanopod.nanopod.treezh.cn/v1
kind: NanoPod
metadata:
  name: my-nano-pod
  namespace: nano-pod-test
spec:
  template:
    spec:
      containers:
      - name: mysql01
        env:
          - name: "env0101"
            value: "value0103"
          - name: "env0102"
            value: "value0103"
        resources:
          limits:
            cpu: 500m
            memory: 128Mi
          requests:
            cpu: 10m
            memory: 64Mi
```

The structure of spec.template is exactly the same as of deployment.spec.template.

Make sure the NanoPod had been created.

```shell
kubectl get nanopod -n nano-pod-test

NAME          AGE
my-nano-pod   12s
```

Since then, the pods created with label `nano-pods: "my-nano-pod"` would be automatically patched by NanoPod `my-nano-pod`.

If you apply a Deployment like this:

```yaml
# mysql.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mysql-nano-pod-enabled
  namespace: nano-pod-test
  labels:
    app: nginx
spec:
  replicas: 1
  selector:
    matchLabels:
      app:  nginx-nano-pod-enabled
  template:
    metadata:
      labels:
        app:  nginx-nano-pod-enabled
        nano-pods: ""
    spec:
      containers:
      - name: mysql01
        image: treezh-docker.pkg.coding.net/demo03/public/mysql:8.0.31
        env:
          - name: "env0102"
            value: "value0102"
          - name: "MYSQL_ALLOW_EMPTY_PASSWORD"
            value: "true"
          - name: "MYSQL_DATABASE"
            value: "mydb"
        livenessProbe:
          tcpSocket:
            port: 3306
          initialDelaySeconds: 15
          periodSeconds: 20
        readinessProbe:
          tcpSocket:
            port: 3306
          initialDelaySeconds: 5
          periodSeconds: 10

```

Run the following script to get the pod.
 
```shell
kubectl get pod -l app=mysql-nano-pod-enabled -o yaml
```

you will get a pod with NanoPod patched.

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    app: mysql-nano-pod-enabled
    nano-pods: my-nano-pod
  name: mysql-nano-pod-enabled-869c5bd95b-bkd4p
  namespace: nano-pod-test
  uid: 2a34b39c-9e55-4158-aa83-16db56009f40
spec:
  containers:
    - env:
        - name: env0101 # this env is added by NanoPod
          value: value0103
        - name: env0102
          value: value0103 # this value is overwritten by NanoPod
        - name: MYSQL_ALLOW_EMPTY_PASSWORD
          value: "true"
        - name: MYSQL_DATABASE
          value: mydb
      image: treezh-docker.pkg.coding.net/demo03/public/mysql:8.0.31
      imagePullPolicy: IfNotPresent
      livenessProbe:
        failureThreshold: 3
        initialDelaySeconds: 15
        periodSeconds: 20
        successThreshold: 1
        tcpSocket:
          port: 3306
        timeoutSeconds: 1
      name: mysql01
      readinessProbe:
        failureThreshold: 3
        initialDelaySeconds: 5
        periodSeconds: 10
        successThreshold: 1
        tcpSocket:
          port: 3306
        timeoutSeconds: 1
      resources: # resources are added by NanoPod 
        limits:
          cpu: 500m
          memory: 512Mi
        requests:
          cpu: 100m
          memory: 256Mi
```

Then you can see env `env0101` and resources are added, the value of env `env0102` is overwritten from `value0102` to `value0103`.

## Contributing

 If you have any problem, welcome to create issue or PR to this repo ,or contact me by email `duncan.tree.zhou@gmail.com`.

## License

[Apache 2.0 License](./LICENSE).

