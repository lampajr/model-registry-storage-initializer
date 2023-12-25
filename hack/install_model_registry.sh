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

namespace=model-registry
while getopts ":hn:" option; do
   case $option in
      h) # display Help
         Help
         exit;;
      r) # override namespace
          namespace=$OPTARG;;
     \?) # Invalid option
         echo "Error: Invalid option"
         exit;;
   esac
done

# Install model registry operator
kubectl apply -k "https://github.com/opendatahub-io/model-registry-operator.git/config/default?ref=main"

# Install model registry CR and Postgres database
if ! kubectl get namespace "$namespace" &> /dev/null; then
   kubectl create namespace $namespace
fi
kubectl -n $namespace apply -k "https://github.com/opendatahub-io/model-registry-operator.git/config/samples?ref=main"

# Wait for model registry deployment
condition="false"
while [ "${condition}" != "True" ]
do
  sleep 5
  condition="`kubectl -n $namespace get mr modelregistry-sample --output=jsonpath='{.status.conditions[?(@.type=="Available")].status}'`"
done