#!/bin/bash

mdrip=$1
mdDir=testdata

echo "Creating ${mdDir}"
${mdrip} writemd ${mdDir}

${mdrip} test ${mdDir}

echo "Deleting ${mdDir}"
rm -rf ${mdDir}
