import asyncio
import secrets  # Use secrets module for cryptographically secure random numbers

from examples.bulk_fanout.worker import ParentInput, bulk_parent_wf
from hatchet_sdk import Hatchet
from hatchet_sdk.clients.admin import TriggerWorkflowOptions


async def main() -> None:
    hatchet = Hatchet()

    # Generate a random stream key to use to track all
    # stream events for this workflow run.

    streamKey = "streamKey"
    streamVal = f"sk-{secrets.randbelow(100) + 1}"  # Cryptographically secure random number

    # Specify the stream key as additional metadata
    # when running the workflow.

    # This key gets propagated to all child workflows
    # and can have an arbitrary property name.
    bulk_parent_wf.run(
        input=ParentInput(n=2),
        options=TriggerWorkflowOptions(additional_metadata={streamKey: streamVal}),
    )

    # Stream all events for the additional meta key value
    listener = hatchet.listener.stream_by_additional_metadata(streamKey, streamVal)

    async for event in listener:
        print(event.type, event.payload)


if __name__ == "__main__":
    asyncio.run(main())