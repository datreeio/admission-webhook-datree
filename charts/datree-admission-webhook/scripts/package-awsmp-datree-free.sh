# This script is used to package the helm chart

echo "Let's package the helm chart! 📦"

echo "📦 Creating values.yaml file..."

cratedValuesFilePath="./charts/datree-admission-webhook/values.yaml"
# create values file with the values from the values-awsmp-datree.yaml file
yq e ./charts/datree-admission-webhook/values-awsmp-datree-free.yaml  >> ${cratedValuesFilePath}

# wait for values.yaml to be created
while [ ! –e ${path} ]
do
  sleep 1
done


echo "📦 Packaging helm chart..."
# package the chart with the image tag from the created values.yaml file as version
version=$(yq e '.image.tag' ${cratedValuesFilePath})
echo $version
helm package ./charts/datree-admission-webhook --destination . --version ${version}

# remove created values file
echo "📦 Removing values.yaml file..."
rm ${cratedValuesFilePath}

echo "✅ All done! 🎉"


