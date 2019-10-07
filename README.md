<!-- PROJECT LOGO -->  
<br />  
<p align="center">  
  <a href="https://github.com/github_username/repo">  
    <img src="docs/assets/logo.png" alt="Logo" width="80" height="80">  
  </a>  
  <p align="center">  
    Fogbow metrics for any scraper  
    <br />  
    <a href="[https://www.fogbowcloud.org/](https://www.fogbowcloud.org/)"><strong>Explore the docs »</strong></a>  
    <br />  
    <br />  
    <a href="https://github.com/fogbow">Open Source</a>    
  </p>  
</p>  
  
<!-- TABLE OF CONTENTS -->  
## Table of Contents  
  
* [About the Project](#about-the-project)  
  * [Built With](#built-with)  
* [Getting Started](#getting-started)  
  * [Prerequisites](#prerequisites)  
  * [Installation](#installation)  
* [Usage](#usage)  
* [Roadmap](#roadmap)  
* [Contributing](#contributing)  
* [License](#license)  
* [Contact](#contact)  
* [Acknowledgements](#acknowledgements)  
  
  
  
<!-- ABOUT THE PROJECT -->  
## About The Project  
  
The telemetry aggregator is intended as a flexible monitoring middleware that offers support for platforms such as Prometheus (for now, the only supported) that can centralize information from Fogbow instances and even federations.

### Built With  
  
* [Golang]([https://github.com/golang/go](https://github.com/golang/go))  
* [Mongodb]([https://github.com/mongodb/mongo](https://github.com/mongodb/mongo))  
* [Docker tools (swarm and so on)]([https://github.com/docker](https://github.com/docker))  
* [Remote deployment provided by Ansible]([https://github.com/ansible/ansible](https://github.com/ansible/ansible))
    
<!-- GETTING STARTED -->  
## Getting Started  
  
To get a local copy up and running follow these simple steps.   
  
You can simply deploy a stack of the services required to run the Aggregator only by running the [deploy script]([https://github.com/emanueljoivo/aggregator/blob/master/ansible/stack/deploy-stack.sh](https://github.com/emanueljoivo/aggregator/blob/master/ansible/stack/deploy-stack.sh)) with root privileges.    
```sh  
sh deploy-stack.sh
```  
### Remote deployment  
  1. Clone the repo  
```sh  
git clone https:://github.com/emanueljoivo/aggregator.git  
```  
2. Enter to the ansible folder and fill the fields on the hosts.conf file
```sh  
cd ansible/ && vim hosts.conf  
```  
3. Once the information in the file is correct, execute the follow command with root privileges
```sh  
sh install.sh
```  
  
  
  
<!-- USAGE EXAMPLES -->  
## Usage  
  
The API tends to be REST and has only one endpoint with two methods
```
POST /metric
GET  /metric/{id}
```  
where {id} is the id of the metric generated when it was generated.
```json
{
   "name":"metric_name",
   "value": 0,
   "timestamp":1568834244853,
   "help":"Some short description about the metric.",
   "metadata":{
      "service": "some_service"
   }   
}
```

<!-- ROADMAP -->  
## Roadmap  
  
See the [open issues](https://github.com/github_username/repo/issues) for a list of proposed features (and known issues).  
  
  
  
<!-- CONTRIBUTING -->  
## Contributing  
  
Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.  
  
1. Fork the Project  
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)  
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)  
4. Push to the Branch (`git push origin feature/AmazingFeature`)  
5. Open a Pull Request  
  
<!-- LICENSE -->  
## License  
  
Distributed under the MIT License. See `LICENSE` for more information.  
  
<!-- CONTACT -->  
## Contact  
  
Emanuel Joívo - [@emanueljoivo](https://twitter.com/emanueljoivo) - emanueljoivo@gmail.com
  
Project Link: [https://github.com/emanueljoivo/aggregator]([https://github.com/emanueljoivo/aggregator](https://github.com/emanueljoivo/aggregator))