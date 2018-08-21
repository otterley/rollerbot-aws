[![CircleCI](https://circleci.com/gh/otterley/rollerbot-aws/tree/master.svg?style=svg)](https://circleci.com/gh/otterley/rollerbot-aws/tree/master)

# RollerBot

## Purpose

RollerBot is a system for automating the replacement of instances in an AWS EC2
Auto Scaling Group. If you need the ability to update the AMI on an existing
group without downtime, with low cost, and with little performance impact (at
the price of more transition latency), RollerBot may be right for you.

RollerBot is especially useful for:

* Updating ECS cluster instances
* Updating instances that host stateful services such as Kafka and Consul

## How it works

RollerBot has several different implementations, depending on your cluster's needs.

### Grow method

The "grow" method is best for stateless Auto Scaling Groups.  It leverages your 
existing Scaling Policy to achieve a slow update of the instances.  It works
by increasing the Auto Scaling Group's Desired Count, then allowing the existing
Scaling Policy to decrease the number of instances back to the steady-state count.
Once all instances have been replaced, the process is complete.

### Replace-after method

The "replace-after" method is best for stateful Auto Scaling Groups whose state
is replicated on multiple instances in the cluster (for example, Kafka and Consul).
It works by decreasing the Auto Scaling Group's Desired Count by one, waiting for 
an instance to cleanly terminate, then increasing the Desired Count by one again, 
to return the Group to its steady-state count.  

## Installation

TBD

## Notes

### CloudTrail
CloudTrail Logs **must** be enabled on the AWS account in which RollerBot is
used so that it can detect when the Auto Scaling Group's Launch Configuration
has been updated. 

## License

Apache 2.0 licensed.
