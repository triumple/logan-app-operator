oc delete -f examples/test-java.yaml --ignore-not-found=true
start=$(date +%s.%N)
start_s=$(echo $start | cut -d '.' -f 1)
start_ns=$(echo $start | cut -d '.' -f 2)
oc create -f examples/test-java.yaml
end=$(date +%s.%N)
end_s=$(echo $end | cut -d '.' -f 1)
end_ns=$(echo $end | cut -d '.' -f 2)

time=$((($end_s - $start_s) * 1000 + ($end_ns / 1000000 - $start_ns / 1000000)))
echo "$time ms"
