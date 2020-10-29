# Template-Service-Broker

> Template-Service-Broker for HyperCloud Service

## prerequisite Install
- Template-Operator
- CatalogController

## Install Cluster-Template-Service-Broker
> 사용자가 공통으로 사용하는 ClusterTemplate 서비스를 제공하기 위한 Broker 입니다.
1. Cluster-Template-Service-Broker를 설치하기 위한 네임스페이스를 생성 합니다.
    - kubectl create namespace {YOUR_NAMESPACE}
2. 아래의 command로 Cluster-Template-Service-Broker를 생성 합니다.
    - kubectl apply -f cluster_tsb.yaml -n {YOUR_NAMESPACE} ([파일](./deploy/cluster_tsb.yaml))
    - 비고: 단, 파일 내부 ClusterRoleBinding의 namespace를 {YOUR_NAMESPACE}로 변경 해야 합니다.
    - 비고: deployment 내부의 image 경로는 사용자 환경에 맞게 수정 해야 합니다.

---

## Install Namespaced-Template-Service-Broker
> 사용자가 직접 만든 Template 서비스를 제공하기 위한 Broker 입니다.
>> 비고: Template 생성한 네임스페이스에 Broker를 생성 해야 합니다.
1. 아래의 command로 Namespaced-Template-Service-Broker를 생성 합니다.
    - kubectl apply -f tsb.yaml -n {YOUR_TEMPLATE_NAMESPACE} ([파일](./deploy/tsb.yaml))
    - 비고: deployment 내부의 image 경로는 사용자 환경에 맞게 수정 해야 합니다.

## Test
```shell
$ curl -X GET http://{SERVER_IP}:{SERVER_PORT}/v2/catalog
```

## ServiceBroker 등록
1. Cluster-Service-Broker
    ```yaml
    apiVersion: servicecatalog.k8s.io/v1beta1
    kind: ClusterServiceBroker
    metadata:
      name: hyperbroker4
    spec:
      url: 'http://{SERVER_IP}:{SERVER_PORT}'
    ```
2. Namespace-Service-Broker
    ```yaml
    apiVersion: servicecatalog.k8s.io/v1beta1
    kind: ServiceBroker
    metadata:
      name: hyperbroker4
      namespace: {YOUR_TEMPLATE_NAMESPACE}
    spec:
      url: 'http://{SERVER_IP}:{SERVER_PORT}'
    ```