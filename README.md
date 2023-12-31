# Atlas

Atlas is a loadbalancer built in GoLang.
A load balancer is a device that distributes network or application traffic across different servers. It acts as a reverse proxy to improve the concurrent user capacity and overall reliability of applications.

## Features

- Can Handle concurrent requests (All credits goes to Goroutines and mutex)
- Works as Reverse Proxy (used httputil.ReverseProxy)
- Critical logs avaialable (used uber's zap package)
- Supports Weighted and normal RoundRobin and Least-Connection Algorithm

## How to Run it in your local environtment

- clone the repo

  ```
   git clone https://github.com/IRSHIT033/Atlas.git
  ```

- Go to the root directory of the project
- Create a **config.yaml** file
- Configure it

  ```
  ###-----------config.yaml------------###

  ## Load Balancer Port
  loadbalance_port: 3333

  ## Retry limit of a backend server
  max_attempt_limit: 1

  ## Backend servers host
  backends:
    - url: "http://localhost:4001"
      weight: 1
    - url: "http://localhost:4002"
      weight: 2
  # put url and the respective weights your servers

  ## Strategy
  strategy: "round-robin"
  ## u can put any one of the two strategies below
  # least-connection or
  # round-robin

  ```

- Run the project
  ```
   go run main.go
  ```
  or ( if you want to run it in docker )
  ```
  docker compose up
  ```
