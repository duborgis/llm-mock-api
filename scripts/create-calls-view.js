// Creates/replaces the "calls" view in the openmeter_mock database, joining each
// raw_responses document (LiteLLM's original provider-shaped response, captured by
// litellm/custom_callback.py) with the OpenMeter event closest to it in time for the same
// model (captured by the LiteLLM "openmeter" callback).
//
// The two collections don't share a key: raw_responses is keyed by LiteLLM's internal call
// id, while events is keyed by the LLM response id (falling back to the call id only when the
// response has none) — so the join here is a best-effort match on model + a 5s time window
// rather than an exact foreign key. Run via `make mongo-view` (see Makefile).
db.calls.drop();
db.createView("calls", "raw_responses", [
  {
    $lookup: {
      from: "events",
      let: { model: "$model", ts: "$received_at" },
      pipeline: [
        {
          $match: {
            $expr: {
              $and: [
                { $eq: ["$data.model", "$$model"] },
                { $lte: [{ $abs: { $subtract: ["$received_at", "$$ts"] } }, 5000] },
              ],
            },
          },
        },
        { $sort: { received_at: 1 } },
        { $limit: 1 },
      ],
      as: "event",
    },
  },
  { $unwind: { path: "$event", preserveNullAndEmptyArrays: true } },
  {
    $project: {
      _id: 0,
      call_id: "$_id",
      model: 1,
      route: 1,
      cache_hit: 1,
      received_at: 1,
      raw_response: 1,
      event_id: "$event._id",
      event_type: "$event.type",
      event_subject: "$event.subject",
      event_data: "$event.data",
      event_time: "$event.time",
    },
  },
]);
print("openmeter_mock.calls view created");
