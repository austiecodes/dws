import { get, post } from "./api";
import type { Container } from "../types";

export async function fetchContainers(): Promise<Container[]> {
  const response = await get<{ containers: Container[] }>("/api/v1/containers");
  return response.containers;
}

export async function fetchContainerImages(): Promise<string[]> {
  const response = await get<{ images: string[] }>("/api/v1/containers/images");
  return response.images;
}

export async function createContainer(
  image: string,
  password?: string
): Promise<Container> {
  const body: { image: string; password?: string } = { image };
  if (password) {
    body.password = password;
  }
  const response = await post<{ container: Container }>("/api/v1/containers", body);
  return response.container;
}

export async function stopContainer(uuid: string): Promise<void> {
  await post<{ message: string }>(`/api/v1/containers/${uuid}/stop`, {});
}

export async function startContainer(uuid: string): Promise<void> {
  await post<{ message: string }>(`/api/v1/containers/${uuid}/start`, {});
}

export async function deleteContainer(uuid: string): Promise<void> {
  const response = await fetch(`/api/v1/containers/${uuid}`, {
    method: "DELETE",
    credentials: "include",
    headers: {
      Accept: "application/json",
    },
  });
  if (!response.ok) {
    const text = await response.text();
    const data = text ? JSON.parse(text) : {};
    throw new Error(data.error || "Delete failed");
  }
}
