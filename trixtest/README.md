# trixtest

this directory contains the Earthfile to build images for the betch/trixtest dockerhub repo

the trixtest image contains a matrix installation that includes two users: trix (admin) and bot (non-admin)

this image is used for running the integration tests in this repo

to run: `earthly --push +all`
(you need to be logged into docker as the betch user to push images to dockerhub)
