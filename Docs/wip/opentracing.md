# OpenTracing

## Development and Testing Usage

**Local**
```bash
docker run -d -p5775:5775/udp -p6831:6831/udp -p6832:6832/udp   -p5778:5778 -p16686:16686 -p14268:14268 -p9411:9411 jaegertracing/all-in-one:0.8.0
```

**Kubernetes**
```bash
kubectl create --namespace fission -f https://raw.githubusercontent.com/jaegertracing/jaeger-kubernetes/master/all-in-one/jaeger-all-in-one-template.yml
```

Once deployed, navigate to http://localhost:16686 to view the workflow traces