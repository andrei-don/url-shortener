# url-shortener

The details of the app architecture and deployment are available at https://alexstan.cloud/posts/apps/url-shortener/.

## Installing the helm release

The app has dependencies on [redis-operator](https://github.com/OT-CONTAINER-KIT/redis-operator) and [cloudnative-pg](https://github.com/cloudnative-pg/cloudnative-pg) helm charts. 

Before installing the helm release, install the operators with the below commands:

```shell
helm repo add ot-helm https://ot-container-kit.github.io/helm-charts/
helm repo add cnpg https://cloudnative-pg.github.io/charts
helm install redis-operator ot-helm/redis-operator --namespace ot-operators --create-namespace
helm upgrade --install cnpg --namespace cnpg-system --create-namespace cnpg/cloudnative-pg
```

The application helm release can be installed by navigating to /charts/url-shortener and running:

```shell
helm install url-shortener .
```