for i in {1..100}; do
    echo "Loop iteration: $i"
    oc create ns tzlil-test-bash-$i
done
