package rhsm

const helpAWS10 = `
This system appears to be running on Amazon Web Services.

Please see the following Red Hat documentation for instructions on how to subscribe your RHEL 10 system on Amazon Web Services:

  RHEL 10 on Amazon Web Services:
    https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/deploying_and_managing_rhel_on_amazon_web_services/index

  Attaching Red Hat subscriptions on Amazon Web Services:
    https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/deploying_and_managing_rhel_on_amazon_web_services/deploying-a-rhel-image-as-an-ec2-instance-on-aws

Alternatively, you can register this system directly with Red Hat by using the RHC client:

  sudo rhc connect --activation-key=<activation_key> --organization=<organization_ID>

For detailed RHEL 10 registration instructions:
  https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/automatically_installing_rhel/registering-your-rhel-system
`
