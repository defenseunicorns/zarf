import { Capability, a, Log } from "pepr";

/**
 *  The Webhook Capability is an example capability to demonstrate using webhooks to interact with Zarf package deployments.
 *  To test this capability you run `pepr dev`and then deploy a zarf package!
 */
export const Webhook = new Capability({
  name: "example-webhook",
  description:
    "A simple example capability to show how webhooks work with Zarf package deployments.",
  namespaces: ["zarf"],
});

const { When } = Webhook;

When(a.Secret)
  .IsCreatedOrUpdated()
  .InNamespace("zarf")
  .WithLabel("package-deploy-info")
  .Mutate(async request => {
    const secret = request.Raw;
    let secretData;
    let secretString: string;
    let manuallyDecoded = false;

    // Pepr does not decode/encode non-ASCII characters in secret data: https://github.com/defenseunicorns/pepr/issues/219
    try {
      secretString = atob(secret.data.data);
      manuallyDecoded = true;
    } catch (err) {
      secretString = secret.data.data;
    }

    // Parse the secret object
    try {
      secretData = JSON.parse(secretString);
    } catch (err) {
      throw new Error(`Failed to parse the secret.data.data: ${err}`);
    }

    for (const deployedComponent of secretData?.deployedComponents ?? []) {
      if (deployedComponent.status === "Deploying") {
        Log.info(
          `The component ${deployedComponent.name} is currently deploying`,
        );

        const componentWebhook =
          secretData.componentWebhooks?.[deployedComponent?.name]?.[
            "test-webhook"
          ];

        // Check if the component has a webhook running for the current package generation
        if (componentWebhook?.observedGeneration === secretData.generation) {
          Log.debug(
            `The component ${deployedComponent.name} has already had a webhook executed for it. Not executing another.`,
          );
        } else {
          // Seed the componentWebhooks map/object
          if (!secretData.componentWebhooks) {
            secretData.componentWebhooks = {};
          }

          // Update the secret noting that the webhook is running for this component
          secretData.componentWebhooks[deployedComponent.name] = {
            "test-webhook": {
              name: "test-webhook",
              status: "Running",
              observedGeneration: secretData.generation,
            },
          };

          try {
            await sleep(10);
            secretData.componentWebhooks[deployedComponent?.name][
              "test-webhook"
            ].status = "Succeeded";
          } catch (err) {
            secretData.componentWebhooks[deployedComponent?.name][
              "test-webhook"
            ].status = "Failed";
            Log.error(`Error sleeping: ${err}`);
          }
        }
      }
    }

    if (manuallyDecoded === true) {
      secret.data.data = btoa(JSON.stringify(secretData));
    } else {
      secret.data.data = JSON.stringify(secretData);
    }
  });

function sleep(seconds: number) {
  return new Promise(resolve => setTimeout(resolve, seconds * 1000));
}
