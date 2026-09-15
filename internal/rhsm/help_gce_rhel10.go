package rhsm

const helpGCE10 = `
This system appears to be running on Google Cloud.

Please see the following Red Hat documentation for instructions on how to subscribe your RHEL 10 system on Google Cloud:

  RHEL 10 on Google Cloud:
    https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/deploying_and_managing_rhel_on_google_cloud/index

  Attaching Red Hat subscriptions on Google Cloud:
    https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/deploying_and_managing_rhel_on_google_cloud/deploying-a-rhel-image-as-a-google-compute-engine-instance-on-gcp

Alternatively, you can register this system directly with Red Hat by using the RHC client:

  sudo rhc connect --activation-key=<activation_key> --organization=<organization_ID>

For detailed RHEL 10 registration instructions:
  https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/automatically_installing_rhel/registering-your-rhel-system
`
