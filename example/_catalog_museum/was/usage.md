## 기본 제공 Pipeline Template
* WAS  
TemplateInstance 생성 시 CI/CD를 위한 Tekton Pipeline을 생성하며, 초기 WAS Deployment를 위한 PipelineRun을 구동함
  * [Apache](./apache/apache-template.yaml)
  * [Django](./django/django-template.yaml)
  * [Node.js](./nodejs/nodejs-template.yaml)
  * [Tomcat](./tomcat/tomcat-template.yaml)
  * [Wildfly](./wildfly/wildfly-template.yaml)

* WAS + DB  
TemplateInstance 생성 시 DB Deployment 및 CI/CD를 위한 Tekton Pipeline을 생성하며, 초기 WAS Deployment를 위한 PipelineRun을 구동함
  * [Node.js+MySQL](../was-db/nodejs-mysql-template.yaml)

각 템플릿의 `objects[0].tasks[0].params[0].value`(BUILDER_IMAGE) 각 환경의 레지스트리로 주소 수정 필요  
각 템플릿을 `kubectl` 또는 Hypercloud Dashboard에서 추가

### Template 구성
* Service for WAS Deployment
* ConfigMap for WAS Deployment
* CI/CD PipelineResource
* CI/CD Pipeline
* CI/CD PipelineRun
* (WAS+DB Only) Secret for DB Connection (User ID/PW)
* (WAS+DB Only) Service for DB Deployment
* (WAS+DB Only) DB Deployment

### Template Input
* Parameter
    * `PIPELINE_NAME`: 생성될 파이프라인 이름
    * `NAMESPACE`: 파이프라인이 생성될 네임스페이스
    * `SERVICE_ACCOUNT_NAME`: CI/CD에 사용될 서비스 어카운트 이름
    * `WAS_PORT`: WAS 접근에 사용하는 포트 번호
    * `GIT_URL`: 소스 코드 Git 주소
    * `GIT_REV`: 소스 코드 Git Revision
    * `IMAGE_URL`: WAS+어플리케이션 이미지가 저장될 주소

### CI/CD Pipeline 구성
1. S2I(Source-to-Image) Task
    1. S2I Build  
: Source &rightarrow; Dockerfile.gen
    2. Buildah bud  
: Dockerfile.gen &rightarrow; Application Image
    3. Buildah push  
: Application Image &rightarrow; Remote Registry
2. Scan Task
    1. Scan Image
: Klar+Clair 이용한 이미지 취약점 분석
3. Deploy Task
    1. Create YAML  
: `deployment.yaml` 파일 생성
    2. Kubectl  
: `deployment.yaml` 파일을 이용한 Deploy

### CI/CD Pipeline Input
* Resource
    * source-repo (PipelineResource-git): 어플리케이션 소스 Git 주소
    * image (PipelineReource-image): 어플리케이션 이미지가 저장될 주소
* Parameter
    * app-name: Deployment 이름
    * replica: Replica 수
    * port: 공개 Port 번호

## 예시
아래 예시는 Tomcat 어플리케이션에 대한 CI/CD 예시에 해당하며, 5가지 기본 제공 WAS에 대한 예시는 [링크](https://github.com/tmax-cloud/hypercloud-operator/tree/master/_catalog_museum/was) 에서 모두 확인 가능

1. `ServiceAccount`, `ClusterRole`, `ClusterRoleBinding` 생성 (Deployment 단계에서 사용)
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: tutorial-service
  namespace: default
```
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: tutorial-role
  namespace: default
rules:
  - apiGroups:
      - apps
    resources:
      - deployments
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
  - apiGroups:
      - ""
    resources:
      - configmaps
      - secrets
    verbs:
      - get
      - list
      - watch
      - create
      - update
      - patch
      - delete
```
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: tutorial-binding
  namespace: default
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: tutorial-role
subjects:
  - kind: ServiceAccount
    name: tutorial-service
    namespace: default
```

2. `TemplateInstance` 생성
```yaml
apiVersion: tmax.io/v1
kind: TemplateInstance
metadata:
  name: tomcat-cicd-template-instance
  namespace: default
spec:
  template:
    metadata:
      name: tomcat-cicd-template
    parameters:
    - name: APP_NAME
      value: tomcat-sample-app
    - name: NAMESPACE
      value: default
    - name: SERVICE_ACCOUNT_NAME
      value: tutorial-service
    - name: GIT_URL
      value: https://github.com/sunghyun_kim3/TomcatMavenApp
    - name: GIT_REV
      value: master
    - name: IMAGE_URL
      value: <레지스트리 주소>/tomcat-sample:latest
    - name: REGISTRY_SECRET_NAME
      value: ''
    - name: WAS_PORT
      value: 8080
    - name: SERVICE_TYPE
      value: LoadBalancer
    - name: PACKAGE_SERVER_URL
      value: ''
    - name: DEPLOY_ENV_JSON
      value: "{}"
```
#### Private Git Repository를 이용할 경우 : [Tekton Authentication Guide](https://github.com/tektoncd/pipeline/blob/master/docs/auth.md#basic-authentication-git) 참조

3. WAS Service 확인
```bash
kubectl -n default get svc tomcat-sample-app-service
```

4.  WAS CI/CD Pipeline 생성 확인
```bash
kubectl -n default get pipeline tomcat-sample-app-pipeline
```

5. PipelineRun 진행사항 확인
```bash
kubectl -n default get pipelinerun -l 'tekton.dev/pipeline=tomcat-sample-app-pipeline'
```

6. PipelineRun 이 성공적으로 완료되면 WAS Deployment 구동 확인
```bash
kubectl -n default get deployment 
```

7. 어플리케이션 코드 수정 시 아래 명령으로 PipelineRun을 실행해 Deployment에 반영
```bash
kubectl get template tomcat-cicd-template -o yaml | \
yq r - 'objects[5]' | \
sed 's/${APP_NAME}/tomcat-sample-app/g' | \
sed 's/${NAMESPACE}/default/g' | \
sed 's/${SERVICE_ACCOUNT_NAME}/tutorial-service/g' | \
kubectl create -f -
```

## 주의 사항
* Tomcat/Wildfly의 경우, S2I Task에서 `WAR_NAME`이라는 환경변수 지정 필요.
Git 소스 최상위 `.s2i/environment` 파일에 `WAR_NAME=<이름>` 명시 ([참고](https://github.com/openshift/source-to-image#build-workflow))
