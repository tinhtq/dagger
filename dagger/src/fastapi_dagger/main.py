import dagger
from dagger import dag, function, object_type
import httpx


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
        source: dagger.Directory,
    ) -> str:
        """Build a Docker image and push it to GHCR.io"""

        client = dagger.Client()

        # Build the Docker image
        built_image = (
            client.container()
            .build(source, dockerfile="./Dockerfile")
            .with_label("version", "1.0.0")  # Optional: Add metadata
        )

        # Push the Docker image
        await built_image.publish(f"{registry}/{image_name}:latest")

        # Return the published image URL
        return f"Image successfully pushed to {registry}/{image_name}:latest"

    @function
    async def scan_and_pr(
        self,
        pull_request_number: str,
        github_repo: str,
        github_token: str,
        source: dagger.Directory,
    ) -> str:
        """
        Scans the application for issues.
        """
        client = dagger.Client()
        container = (
            client.container()
            .from_("python:3.10")
            .with_mounted_directory("/src", source)
            .with_workdir("/src")
            .with_exec(["pip", "install", "-r", "requirements.txt"])  # Install deps
            .with_exec(
                ["sh", "-c", "flake8 app || true"]
            )  # Run flake8 and ignore errors
        )
        scan_results = await container.stdout()
        comment_body = {"body": f"## Scan Results\n\n```\n{scan_results}\n```"}
        headers = {
            "Authorization": f"Bearer {github_token}",
            "X-GitHub-Api-Version": "2022-11-28",
            "Accept": "application/vnd.github+json",
        }
        comment_url = f"https://api.github.com/repos/{github_repo}/issues/{pull_request_number}/comments"

        async with httpx.AsyncClient() as http_client:
            response = await http_client.post(
                comment_url, json=comment_body, headers=headers
            )

        if response.status_code == 201:
            return "Comment posted successfully!"
        else:
            return f"Failed to post comment: {response.text}"
