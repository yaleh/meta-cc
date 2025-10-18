"""
Session Analyzer - Python MCP Server for session data analysis.

This is a minimal MCP server implementation for validating
Go→Python test strategy methodology transfer.
"""

import json
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Any
from dataclasses import dataclass, asdict


@dataclass
class ToolCall:
    """Represents a tool call in a session."""
    tool: str
    timestamp: str
    status: str
    duration_ms: Optional[int] = None

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'ToolCall':
        """Create ToolCall from dictionary."""
        return cls(
            tool=data.get('tool', ''),
            timestamp=data.get('timestamp', ''),
            status=data.get('status', 'success'),
            duration_ms=data.get('duration_ms')
        )


@dataclass
class SessionStats:
    """Session statistics."""
    total_tools: int
    success_count: int
    error_count: int
    unique_tools: int
    avg_duration_ms: float

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return asdict(self)


class SessionParser:
    """Parse session JSONL files."""

    def __init__(self, file_path: str):
        """Initialize parser with file path."""
        self.file_path = Path(file_path)
        if not self.file_path.exists():
            raise FileNotFoundError(f"Session file not found: {file_path}")

    def parse(self) -> List[ToolCall]:
        """Parse JSONL file and return tool calls."""
        tool_calls = []

        try:
            with open(self.file_path, 'r', encoding='utf-8') as f:
                for line_num, line in enumerate(f, 1):
                    line = line.strip()
                    if not line:
                        continue

                    try:
                        data = json.loads(line)
                        if 'tool' in data:
                            tool_calls.append(ToolCall.from_dict(data))
                    except json.JSONDecodeError as e:
                        raise ValueError(
                            f"Invalid JSON at line {line_num}: {e}"
                        )
        except UnicodeDecodeError as e:
            raise ValueError(f"Invalid encoding in file: {e}")

        return tool_calls


class SessionAnalyzer:
    """Analyze session data and provide statistics."""

    def __init__(self, tool_calls: List[ToolCall]):
        """Initialize analyzer with tool calls."""
        self.tool_calls = tool_calls

    def calculate_stats(self) -> SessionStats:
        """Calculate session statistics."""
        if not self.tool_calls:
            return SessionStats(
                total_tools=0,
                success_count=0,
                error_count=0,
                unique_tools=0,
                avg_duration_ms=0.0
            )

        success_count = sum(
            1 for tc in self.tool_calls if tc.status == 'success'
        )
        error_count = sum(
            1 for tc in self.tool_calls if tc.status == 'error'
        )
        unique_tools = len(set(tc.tool for tc in self.tool_calls))

        # Calculate average duration (only for calls with duration)
        durations = [
            tc.duration_ms for tc in self.tool_calls
            if tc.duration_ms is not None
        ]
        avg_duration = sum(durations) / len(durations) if durations else 0.0

        return SessionStats(
            total_tools=len(self.tool_calls),
            success_count=success_count,
            error_count=error_count,
            unique_tools=unique_tools,
            avg_duration_ms=avg_duration
        )

    def query_by_tool(self, tool_name: str) -> List[ToolCall]:
        """Query tool calls by tool name."""
        return [tc for tc in self.tool_calls if tc.tool == tool_name]

    def query_by_status(self, status: str) -> List[ToolCall]:
        """Query tool calls by status."""
        return [tc for tc in self.tool_calls if tc.status == status]

    def get_tool_frequency(self) -> Dict[str, int]:
        """Get frequency count for each tool."""
        frequency = {}
        for tc in self.tool_calls:
            frequency[tc.tool] = frequency.get(tc.tool, 0) + 1
        return frequency


class MCPServer:
    """MCP Server for session analysis."""

    def __init__(self):
        """Initialize MCP server."""
        self.analyzers: Dict[str, SessionAnalyzer] = {}

    def load_session(self, session_id: str, file_path: str) -> None:
        """Load a session from file."""
        parser = SessionParser(file_path)
        tool_calls = parser.parse()
        self.analyzers[session_id] = SessionAnalyzer(tool_calls)

    def get_stats(self, session_id: str) -> Optional[SessionStats]:
        """Get statistics for a session."""
        analyzer = self.analyzers.get(session_id)
        if not analyzer:
            return None
        return analyzer.calculate_stats()

    def query_tools(
        self,
        session_id: str,
        tool: Optional[str] = None,
        status: Optional[str] = None
    ) -> Optional[List[ToolCall]]:
        """Query tool calls with filters."""
        analyzer = self.analyzers.get(session_id)
        if not analyzer:
            return None

        results = analyzer.tool_calls

        if tool:
            results = [tc for tc in results if tc.tool == tool]

        if status:
            results = [tc for tc in results if tc.status == status]

        return results

    def get_tool_frequency(
        self, session_id: str
    ) -> Optional[Dict[str, int]]:
        """Get tool usage frequency for a session."""
        analyzer = self.analyzers.get(session_id)
        if not analyzer:
            return None
        return analyzer.get_tool_frequency()


def main():
    """Main entry point for CLI."""
    import sys

    if len(sys.argv) < 2:
        print("Usage: python session_analyzer.py <session_file>")
        sys.exit(1)

    file_path = sys.argv[1]

    try:
        # Parse session
        parser = SessionParser(file_path)
        tool_calls = parser.parse()

        # Analyze
        analyzer = SessionAnalyzer(tool_calls)
        stats = analyzer.calculate_stats()

        # Print results
        print("Session Statistics:")
        print(f"  Total tool calls: {stats.total_tools}")
        print(f"  Successful: {stats.success_count}")
        print(f"  Errors: {stats.error_count}")
        print(f"  Unique tools: {stats.unique_tools}")
        print(f"  Avg duration: {stats.avg_duration_ms:.2f}ms")

    except FileNotFoundError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == '__main__':
    main()
