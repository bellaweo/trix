# godnd

this directory contains the Earthfile to build images for the betch/godnd dockerhub repo

the godnd image contains golang (the verison is the image tag) plus docker in docker as installed by earthly's dind install script 

to run: `earthly --allow-privileged --push +all`
(you need to be logged into docker as the betch user to push images to dockerhub)
