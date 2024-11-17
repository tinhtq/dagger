import dagger
from dagger import dag, function, object_type


@object_type
class FastapiDagger:
    @function
    def container_echo(self, string_arg: str) -> dagger.Container:
        """Returns a container that echoes whatever string argument is provided"""
        return dag.container().from_("alpine:latest").with_exec(["echo", string_arg])

    @function
    async def grep_dir(self, directory_arg: dagger.Directory, pattern: str) -> str:
        """Returns lines that match a pattern in the files of the provided Directory"""
        return await (
            dag.container()
            .from_("alpine:latest")
            .with_mounted_directory("/mnt", directory_arg)
            .with_workdir("/mnt")
            .with_exec(["grep", "-R", pattern, "."])
            .stdout()
        )

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
            .build(source, dockerfile="./Dockerfile")
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
