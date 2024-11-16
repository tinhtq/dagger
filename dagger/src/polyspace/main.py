import dagger
from dagger import function, object_type


@object_type
class PublishToGHCR:
    @function
    async def build_and_push(
        self,
        registry: str,
        image_name: str,
        username: str,
        password: str,
        source: dagger.Directory = ".",
    ) -> str:
        """Build a Docker image and push it to GHCR.io"""

        client = dagger.Client()

        # Build the Docker image
        built_image = (
            client.container()
            .build(".", dockerfile="./Dockerfile")
            .with_label("version", "1.0.0")  # Optional: Add metadata
        )

        # Log into the Docker registry
        docker_registry = (
            client.container()
            .from_("docker")
            .with_exec(["docker", "login", registry, "-u", username, "-p", password])
        )

        # Push the Docker image
        push_result = await built_image.publish(f"{registry}/{image_name}:latest")

        # Return the published image URL
        return f"Image successfully pushed to {registry}/{image_name}:latest"
