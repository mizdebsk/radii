package rhsm

const helpGeneric10 = `
Please ensure that this system is registered with Red Hat and has access to an active Red Hat Enterprise Linux subscription.

To register this system with Red Hat by using the RHC client:

  sudo rhc connect --activation-key=<activation_key> --organization=<organization_ID>

An activation key and your Red Hat organization ID are required.

For detailed RHEL 10 registration instructions:
  https://docs.redhat.com/en/documentation/red_hat_enterprise_linux/10/html/automatically_installing_rhel/registering-your-rhel-system
`
