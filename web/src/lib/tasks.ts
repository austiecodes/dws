import { get, post } from "./api";
import type { Task, CreateTaskRequest, QueueStatistics } from "../types";

export async function fetchTasks(): Promise<Task[]> {
  const response = await get<{ tasks: Task[] }>("/api/v1/tasks");
  return response.tasks;
}

export async function fetchTask(id: number): Promise<Task> {
  const response = await get<{ task: Task }>(`/api/v1/tasks/${id}`);
  return response.task;
}

export async function createTask(req: CreateTaskRequest): Promise<Task> {
  const response = await post<{ task: Task }>("/api/v1/tasks", req);
  return response.task;
}

export async function cancelTask(id: number): Promise<void> {
  await post<{ message: string }>(`/api/v1/tasks/${id}/cancel`, {});
}

export async function fetchQueueStatistics(): Promise<QueueStatistics> {
  return await get<QueueStatistics>("/api/v1/tasks/queue");
}

