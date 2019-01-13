# Scrimp LB
For a production workload in the cloud, we'd naturally gravitate towards cloud solutions such as the [ELB](https://aws.amazon.com/elasticloadbalancing/pricing/) in AWS, or DigitalOcean's [Load Balancer](https://www.digitalocean.com/products/load-balancer/).

Those are great at what they do, and for the vast majority of prod workloads they're probably the right way to go. Unfortunately, they're quite expensive for a pet project or for a hobby.

At the time of writing (and, importantly, ignoring data transfer costs):

- AWS CLB costs ~$20 a month
- AWS ALB/NLB costs ~$18 a month
- DigitalOcean Load Balancers cost ~$10 a month

Compare that to running cheap instances:

- AWS t3.nano instances start at ~$4.1 a month, or ~$1.2 if using spot instances
- DigitalOcean droplets start at $5 a month

For low-traffic\* deployments it can be economically sensible to run a "load balancer" on an instance and avoid the cost of a more traditional load balancer. On AWS you might even have a NAT instance lying around which can be given a second job.

At higher levels of traffic, the economic reasoning changes and you might end up paying a fair amount in data transfer fees versus a cloud load balancer. Plus, running on an instance is basically always going to be a pain compared to running a managed service.

Another thing to bear in mind is that for a side project, the network and cpu utilization per instance will usually not be high; that is, it seems safe to say most personal projects or toys aren't serving a huge amount of traffic from every instance, and that traffic will generally not be highly intensive. As such, we can take a little CPU and network for maintaining our load-balancing topology quite safely - but that's not to ignore that this _is_ a cost which we must pay.


## Design
A public-facing "load balancer" sits in front of privately-networked "backend" instances which run at least one application.

All instances are networked together in a gossip cluster. Membership of the cluster implies that an instance is either a load balancer or a backend server, although the load balancer might not route to to the backend server without first having received confirmation of what applications are provided by that server via a broadcast exchange.

Ideally, a load balancer is provisioned first (which could be a NAT instance in an AWS VPC for example, or a cheap VPS generally). The load balancer's IP is pushed into some "seed source" (e.g. S3 - but it's easy to write new seed provisioners).

When a backend instance is brought up, a seed provisioner fetches the load balancer's IP from the source, and joins the cluster using the fetched IP.

After joining, a backend server must announce its supported applications to at least one load balancer, which can then share amongst other load balancers as needed. If the load balancer supports the application type, it adds the backend to its list of upstreams (think e.g. how nginx does load balancing) and soft-reloads itself.
