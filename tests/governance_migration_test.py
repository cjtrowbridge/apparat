from pathlib import Path
import unittest


class GovernanceMigrationTest(unittest.TestCase):
    def test_agentic_pipelines_is_the_only_governance_submodule(self):
        modules = Path(".gitmodules").read_text(encoding="utf-8")
        self.assertIn('[submodule "agentic-pipelines"]', modules)
        self.assertNotIn('[submodule "agents"]', modules)
        self.assertFalse(Path("agents").exists())

    def test_deprecated_operational_systems_are_absent(self):
        for path in (
            Path("kanban"),
            Path("downtime"),
            Path("playbooks/how_to_move_kanban_tasks_verbatim.md"),
            Path("playbooks/how_to_use_downtime_to_improve_the_framework.md"),
            Path("references/kanban_verbatim_handling.md"),
            Path("templates/kanban_board.md"),
            Path("templates/downtime_report.md"),
        ):
            self.assertFalse(path.exists(), path)

    def test_root_governance_owns_host_routes_and_build_scope(self):
        agents = Path("AGENTS.md").read_text(encoding="utf-8")
        self.assertIn("## Apparat Task Routing", agents)
        self.assertIn("## Apparat Build Pipeline Scope", agents)
        self.assertIn("agentic-pipelines/AGENTS.md", agents)
        for obsolete in ("./agents/", "kanban", "downtime"):
            self.assertNotIn(obsolete, agents.lower())


if __name__ == "__main__":
    unittest.main()
