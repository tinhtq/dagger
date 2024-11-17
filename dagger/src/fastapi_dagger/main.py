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
        source: dagger.Directory,
        github_repo: str,
        github_token: str,
        base_branch: str = "main",
        pr_branch: str = "scan-fix",
    ) -> str:
        """
        Scans the application for issues and creates a PR with the scan results.
        """

        async with dagger.Connection() as client:
            # Step 1: Prepare the container for scanning
            container = (
                client.container()
                .from_("python:3.10")
                .with_mounted_directory("/src", source)
                .with_workdir("/src")
                .with_exec(["pip", "install", "-r", "requirements.txt"])  # Install deps
                .with_exec(["flake8", "."])  # Run static analysis (e.g., Flake8)
            )

            # Step 2: Capture scan results
            scan_results = await container.stdout()

            # Step 3: Push changes to a new branch (assuming no fixes are made)
            git_container = (
                client.container()
                .from_("alpine/git")
                .with_exec(["git", "config", "--global", "user.name", "Dagger Bot"])
                .with_exec(["git", "config", "--global", "user.email", "bot@dagger.io"])
                .with_mounted_directory("/src", source)
                .with_workdir("/src")
                .with_exec(["git", "checkout", "-b", pr_branch])
                .with_exec(["git", "add", "."])
                .with_exec(["git", "commit", "-m", "Add scan results"])
                .with_exec(["git", "push", "origin", pr_branch])
            )

            await git_container.stdout()

            # Step 4: Create a Pull Request using GitHub API
            pr_data = {
                "title": "Scan Results: Fix Issues",
                "head": pr_branch,
                "base": base_branch,
                "body": f"## Scan Results\n\n```\n{scan_results}\n```",
            }

            headers = {"Authorization": f"Bearer {github_token}"}
            pr_url = f"https://api.github.com/repos/{github_repo}/pulls"

            async with httpx.AsyncClient() as http_client:
                response = await http_client.post(pr_url, json=pr_data, headers=headers)

            if response.status_code == 201:
                return "Pull Request created successfully!"
            else:
                return f"Failed to create PR: {response.text}"
