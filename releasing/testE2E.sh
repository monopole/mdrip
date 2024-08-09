#!/bin/bash

mdrip=$1
mdDir=testdata

echo "Creating ${mdDir}"
${mdrip} generatetestdata ${mdDir}

${mdrip} test ${mdDir}

echo "Deleting ${mdDir}"
rm -rf ${mdDir}
