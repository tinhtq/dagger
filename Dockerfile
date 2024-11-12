FROM ubuntu:20.04

RUN apt-get update && apt-get install -y \
    build-essential \
    wget

# PolySpace installation (mocked here as an example)
COPY PolySpaceInstaller.sh /opt/
RUN chmod +x /opt/PolySpaceInstaller.sh && /opt/PolySpaceInstaller.sh

ENV POLYSPACE_HOME /opt/polyspace
ENV PATH "$POLYSPACE_HOME/bin:$PATH"

COPY scripts/run_polyspace.sh /usr/local/bin/run_polyspace.sh
RUN chmod +x /usr/local/bin/run_polyspace.sh

ENTRYPOINT ["run_polyspace.sh"]
