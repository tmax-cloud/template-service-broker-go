# CI/CD Trigger Template Guide

CI/CD Trigger Template은 WAS CI/CD Template과 함께 쓰는 템플릿으로, GitLab Push 이벤트 발생 시 자동으로 CI/CD 파이프라인이 수행되도록 해주는 Tekton Trigger 오브젝트를 포함함.

## Step.0 - Prerequisites
1. Tekton Pipeline 및 Tekton Trigger 설치  
: [링크](https://github.com/tmax-cloud/hypercloud-install-guide/blob/master/Tekton_CI_CD/trigger.md) 참조

2. ServiceAccount/Role/RoleBinding 생성
```yaml
kind: ServiceAccount
apiVersion: v1
metadata:
  name: tekton-triggers-example-sa
  namespace: default
```
```yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tekton-triggers-example-minimal
  namespace: default
rules:
  - verbs:
      - get
    apiGroups:
      - triggers.tekton.dev
    resources:
      - eventlisteners
      - triggerbindings
      - triggertemplates
  - verbs:
      - get
      - list
      - watch
    apiGroups:
      - ''
    resources:
      - configmaps
      - secrets
  - verbs:
      - create
    apiGroups:
      - tekton.dev
    resources:
      - pipelineruns
      - pipelineresources
      - taskruns
```
```yaml
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: tekton-triggers-example-binding
  namespace: default
subjects:
  - kind: ServiceAccount
    name: tekton-triggers-example-sa
    namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: tekton-triggers-example-minimal
```

## Step.1 - WAS CI/CD Pipeline 생성
1. WAS CI/CD Pipeline Template/TemplateInstance 생성  
(Trigger Template이 아님을 주의)
```bash
kubectl apply -f nodejs-template.yaml
kubectl apply -f nodejs-instance.yaml
```

2. WAS CI/CD PipelineRun 실행 확인 및 완료 대기
```bash
kubectl get pipelinerun -l "tekton.dev/pipeline=nodejs-sample-app-pipeline"
```

3. PipelineRun 완료 이후 Deployment 생성 확인
```bash
kubectl get deployment nodejs-sample-app
```

## Step.2 - EventListener 생성
1. Trigger Template/TemplateInstance 생성
```bash
kubectl apply -f template.yaml
kubectl apply -f instance.yaml
```
* `instance.yaml`의 `APP_NAME` 파라미터 및 `SERVICE_ACCOUNT_NAME` 파라미터는 Step1에서 WAS CI/CD Pipeline을 생성할 때 입력한 `APP_NAME` 파라미터 및 `SERVICE_ACCOUNT_NAME` 파라미터와 동일한 값을 가져야 함.

2. Trigger EventListener 생성 확인
```bash
kubectl get eventlistener nodejs-sample-app-listener
```

## Step.3a - GitLab이 클러스터 내부에 있을 경우

1. EventListener Service의 ClusterIP 확인
```bash
kubectl get svc el-nodejs-sample-app-listener -o jsonpath="{.spec.clusterIP}"
```

## Step.3b - GitLab이 클러스터 외부에 있을 경우

PublicIP의 낭비를 막기 위해 Ingress를 이용해 외부에 노출함.

1. Ingress 생성 또는 편집
```bash
cat << EOF | kubectl apply -f -
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: trigger-el-ingress
  namespace: default
spec:
  rules:
  - http:
      paths:
      - path: /nodejs-trigger
        backend:
          serviceName: el-nodejs-sample-app-listener
          servicePort: 8080
EOF
```

2. Ingress LoadBalancer IP 확인
```bash
kubectl get ingress trigger-el-ingress -o jsonpath="{.status.loadBalancer.ingress[0].ip}"
```

## Step.4 - GitLab 프로젝트에 WebHook 설정

1. (4a-GitLab이 클러스터 내부에 있을 경우)  
`GitLab Admin Area>Settings>Network>Outbound requests` 메뉴의 `Allow requests to the local network from web hooks and services` 항목 체크

2. Step.3a-1 또는 Step.3b-2에서 확인한 IP/URL를 `프로젝트>Settings>Webhooks` 메뉴에 등록

* `URL`:  
(4a-GitLab이 클러스터 내부에 있을 경우): `http://<Step.3a-1 IP>:8080`  
(4b-GitLab이 클러스터 외부에 있을 경우): `http://<Step.3b-2 IP>/nodejs-trigger`
* `SecretToken`: 빈 칸으로 유지
* `Trigger`: `Push events`에 체크
* `SSL verification`: `Enable SSL verification` 체크 해제

## Step.5 - Trigger 작동 확인
1. Git Commit 및 GitLab 프로젝트에 Push

2. 새로운 PipelineRun 생성 확인  
(Step.1-2에서 확인한 PipelineRun이 아닌 새로운 PipelineRun이 생성되어야 함)
```bash
kubectl get pipelinerun -l "tekton.dev/pipeline=nodejs-sample-app-pipeline"
```

3. PipelineRun 완료 이후 Deployment 업데이트 확인
```bash
kubectl get deployment nodejs-sample-app
```

