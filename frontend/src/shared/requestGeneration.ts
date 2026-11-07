// Cleanup must invalidate the latest in-flight generation, not a captured counter value.
export function invalidatePendingRequests(...generations: Array<{ current: number }>) {
  for (const generation of generations) generation.current += 1;
}
