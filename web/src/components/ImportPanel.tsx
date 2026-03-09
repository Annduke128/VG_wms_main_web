import { useCallback, useRef, useState } from "react";
import { api } from "../api/client";
import type { AsyncJob } from "../types/grid";

const FILE_TYPES = [
	{ key: "products", label: "Products" },
	{ key: "inventory", label: "Inventory" },
	{ key: "inbound", label: "Inbound" },
	{ key: "outbound", label: "Outbound" },
] as const;

export function ImportPanel() {
	const [activeJob, setActiveJob] = useState<AsyncJob | null>(null);
	const [uploading, setUploading] = useState(false);
	const fileRefs = useRef<Record<string, HTMLInputElement | null>>({});

	const handleUpload = useCallback(async (fileType: string) => {
		const input = fileRefs.current[fileType];
		if (!input?.files?.[0]) {
			alert("Please select a file first");
			return;
		}

		setUploading(true);
		try {
			const result = (await api.importFile(fileType, input.files[0])) as {
				job_id: string;
			};
			pollJob(result.job_id);
		} catch (err) {
			alert(
				`Upload failed: ${err instanceof Error ? err.message : "Unknown error"}`,
			);
		} finally {
			setUploading(false);
		}
	}, []);

	const pollJob = useCallback(async (jobId: string) => {
		const poll = async () => {
			try {
				const job = (await api.getJob(jobId)) as AsyncJob;
				setActiveJob(job);
				if (job.status === "pending" || job.status === "running") {
					setTimeout(poll, 1000);
				}
			} catch {
				console.error("Poll error");
			}
		};
		poll();
	}, []);

	return (
		<div style={{ padding: 16 }}>
			<h2>Import Excel (.xlsx)</h2>

			<div
				style={{
					display: "grid",
					gridTemplateColumns: "repeat(2, 1fr)",
					gap: 16,
					maxWidth: 600,
				}}
			>
				{FILE_TYPES.map((ft) => (
					<div
						key={ft.key}
						style={{ background: "#f5f5f5", borderRadius: 8, padding: 16 }}
					>
						<h3 style={{ margin: "0 0 8px" }}>{ft.label}</h3>
						<input
							ref={(el) => {
								fileRefs.current[ft.key] = el;
							}}
							type="file"
							accept=".xlsx"
							style={{ marginBottom: 8, display: "block" }}
						/>
						<button
							onClick={() => handleUpload(ft.key)}
							disabled={uploading}
							style={{
								background: "#2196F3",
								color: "#fff",
								border: "none",
								borderRadius: 4,
								padding: "8px 16px",
								cursor: uploading ? "not-allowed" : "pointer",
							}}
						>
							{uploading ? "Uploading..." : "Upload"}
						</button>
					</div>
				))}
			</div>

			{activeJob && (
				<div
					style={{
						marginTop: 16,
						padding: 16,
						background: "#f0f0f0",
						borderRadius: 8,
					}}
				>
					<h3>Job: {activeJob.job_id.slice(0, 8)}...</h3>
					<p>
						Status:{" "}
						<strong
							style={{
								color:
									activeJob.status === "completed"
										? "green"
										: activeJob.status === "failed"
											? "red"
											: "#888",
							}}
						>
							{activeJob.status}
						</strong>
					</p>
					{activeJob.error && <p style={{ color: "red" }}>{activeJob.error}</p>}
				</div>
			)}
		</div>
	);
}
