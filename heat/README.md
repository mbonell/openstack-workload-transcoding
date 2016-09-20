# Video Transcoding: Orchestration
The Video Transcoding deployment is divided in five templates. Each of them create the resources, install dependencies and run the code needed to up and run all the microservices.
This list represent the order that you have to follow creating the stacks for the deployment.

1. Database
1. Manager
1. Jobs
1. Monitor
1. Worker

## Usage

### Pre-requisites
* OpenStack Heat service enabled on your cloud.
* Images with cloud-init and heat-cfntools packages installed.

### Set the credentials of your cloud to use Heat
```
export OS_AUTH_URL=https://example.com:5000/v3
export OS_PROJECT_NAME=admin
export OS_USERNAME=admin
export OS_PASSWORD=admin
export OS_PROJECT_DOMAIN_ID=default
export OS_USER_DOMAIN_ID=default
```
### Stacks creation
Each template expect different parameters (infra, environment vars and endpoints). Before start launching them, read carefully the parameters seccion inside each HOT.

After the deployment you can validate that all the microservices are running as expected by listing the assigned ports in the instances.
```
netstat -atn
```

  **1. Database**
  ```
  $ cd heat/database
  $ openstack stack create -t database.yaml --parameter key_name=demokey --parameter flavor=m1.small --parameter image=ubuntu-server-14.04 --parameter private_network=internal --parameter volumen_size=1 database
  ```

  **2. Manager**
  ```
  $ cd heat/manager
  $ openstack stack create -t manager.yaml --parameter key_name=demokey --parameter flavor=m1.small --parameter image=ubuntu-server-14.04 --parameter private_network=internal --parameter volumen_size=1 manager
  ```

  **3. Jobs**

  The jobs microservice needs the cloud credentials to interact with the OpenStack services and the database and manager endpoints (IP address).
  ```
  $ cd heat/jobs
  $ openstack stack create -t jobs.yaml --parameter key_name=demokey --parameter flavor=m1.small --parameter image=ubuntu-server-14.04 --parameter private_network=internal --parameter os_auth_url=<OS_AUTH_URL> --parameter os_username=<OS_USERNAME> --parameter os_project_name=<OS_PROJECT_NAME> --parameter os_password=<OS_PASSWORD> --parameter os_domain_id=<OS_PROJECT_DOMAIN_ID> --parameter database_endpoint=<DATABASE_IP> --parameter manager_endpoint=<MANAGER_IP> jobs
  ```

  **4. Monitor**

  The monitor microservice needs the database endpoint (IP address).
  ```
  $ cd heat/monitor
  $ openstack stack create -t monitor.yaml --parameter key_name=demokey --parameter flavor=m1.small --parameter image=ubuntu-server-14.04 --parameter private_network=internal --parameter database_endpoint=<DATABASE_IP> monitor
  ```

  **5. Worker**

  The worker microservice needs the cloud credentials to interact with the OpenStack services and the jobs, manager and monitor endpoints (IP address).
  ```
  $ cd heat/worker
  $ openstack stack create -t worker.yaml --parameter key_name=demokey --parameter flavor=m1.small --parameter image=ubuntu-server-14.04 --parameter private_network=internal --parameter os_auth_url=<OS_AUTH_URL> --parameter os_username=<OS_USERNAME> --parameter os_project_name=<OS_PROJECT_NAME> --parameter os_password=<OS_PASSWORD> --parameter os_domain_id=<OS_PROJECT_DOMAIN_ID> --parameter jobs_endpoint=<JOBS_IP> --parameter manager_endpoint=<MANAGER_IP> --parameter monitor_endpoint=<MONITOR_IP> worker
  ```