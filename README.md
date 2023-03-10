# 🚘 go-share-rides 🚗
Program in Go (language) for students to share rides to Hochschule Fulda. A dummy application not to be used on production directly.


# Setting up (Windows)

## Setting up OSRM backend
1. Install docker
2. Download necessary data from [Geofabrik](http://download.geofabrik.de/) into `./data` folder
3. In the project root folder, run `docker run -t -v "${PWD}/data:/data" ghcr.io/project-osrm/osrm-backend osrm-extract -p /opt/car.lua /data/hessen-latest.osm.pbf` 
4. `docker run -t -v "${PWD}/data:/data" ghcr.io/project-osrm/osrm-backend osrm-partition /data/hessen-latest.osrm`
5. `docker run -t -v "${PWD}/data:/data" ghcr.io/project-osrm/osrm-backend osrm-customize /data/hessen-latest.osrm`
6. `docker run -t -i -p 5000:5000 -v "${PWD}/data:/data" ghcr.io/project-osrm/osrm-backend osrm-routed --algorithm mld /data/hessen-latest.osrm`. This starts up the API server on port 5000.
7. (optional) Start the frontend to view the map `docker run -p 9966:9966 osrm/osrm-frontend`


# Ideation 💡

![image](https://user-images.githubusercontent.com/10389062/213888439-ac570c22-d189-4322-adad-b892e7dd418a.png)
