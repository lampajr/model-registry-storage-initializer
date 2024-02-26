set -e

############################################################
# Help                                                     #
############################################################
Help() {
  # Display Help
  echo "ModelRegistry install script."
  echo
  echo "Syntax: [-n NAMESPACE]"
  echo "options:"
  echo "n Namespace."
  echo
}

namespace=kubeflow
while getopts ":hn:" option; do
   case $option in
      h) # display Help
         Help
         exit;;
      n) # override namespace
          namespace=$OPTARG;;
     \?) # Invalid option
         echo "Error: Invalid option"
         exit;;
   esac
done

# Create namespace if not already existing
if ! kubectl get namespace "$namespace" &> /dev/null; then
   kubectl create namespace $namespace
fi
# Apply model-registry kustomize manifests
kubectl -n $namespace apply -k "https://github.com/kubeflow/model-registry.git/manifests/kustomize/overlays/postgres?ref=main"

# Wait for model registry deployment
modelregistry=$(kubectl get pod -n kubeflow --selector="component=model-registry-server" --output jsonpath='{.items[0].metadata.name}')
kubectl wait --for=condition=Ready pod/$modelregistry -n $namespace --timeout=5m