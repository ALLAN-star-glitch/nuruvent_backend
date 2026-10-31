package domain

// TaskProvisionVideoMeetings is the task type emitted after an event
// is created or updated, instructing the worker to provision video
// meetings for schedules within the rolling 30-day window.
const TaskProvisionVideoMeetings = "video:provision_meetings"