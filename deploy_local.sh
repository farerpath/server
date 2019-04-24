#!/bin/bash
cd ./sessionservice; make release; cd ..
cd ./authservice; make release; cd ..
cd ./albumservice; make release; cd ..
cd ./fileservice; make release; cd ..
cd ./apiservice; make release; cd ..
docker-compose build
docker-compose up
