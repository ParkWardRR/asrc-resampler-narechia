#[cfg(feature = "ebpf")]
pub mod tracing {
    // Stub for eBPF / aya USDT probes
    pub fn trigger_latency_probe(duration_ns: u64) {
        // Fire a statically defined tracepoint
    }
}
