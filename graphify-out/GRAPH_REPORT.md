# Graph Report - blueprint  (2026-04-23)

## Corpus Check
- 28 files · ~177,924 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 412 nodes · 1099 edges · 16 communities detected
- Extraction: 62% EXTRACTED · 38% INFERRED · 0% AMBIGUOUS · INFERRED: 419 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Community 0|Community 0]]
- [[_COMMUNITY_Community 1|Community 1]]
- [[_COMMUNITY_Community 2|Community 2]]
- [[_COMMUNITY_Community 3|Community 3]]
- [[_COMMUNITY_Community 4|Community 4]]
- [[_COMMUNITY_Community 5|Community 5]]
- [[_COMMUNITY_Community 6|Community 6]]
- [[_COMMUNITY_Community 7|Community 7]]
- [[_COMMUNITY_Community 8|Community 8]]
- [[_COMMUNITY_Community 9|Community 9]]
- [[_COMMUNITY_Community 10|Community 10]]
- [[_COMMUNITY_Community 11|Community 11]]
- [[_COMMUNITY_Community 12|Community 12]]
- [[_COMMUNITY_Community 13|Community 13]]
- [[_COMMUNITY_Community 14|Community 14]]
- [[_COMMUNITY_Community 15|Community 15]]

## God Nodes (most connected - your core abstractions)
1. `newTestState()` - 70 edges
2. `newNilLogger()` - 69 edges
3. `writeFile()` - 29 edges
4. `writeBacklogFile()` - 24 edges
5. `initRepo()` - 21 edges
6. `RunRules()` - 18 edges
7. `rule5CommandCheckpoints()` - 18 edges
8. `Open()` - 17 edges
9. `rule3TeamCompliance()` - 15 edges
10. `SyncPlanFile()` - 14 edges

## Surprising Connections (you probably didn't know these)
- `SyncPlanFile()` --calls--> `writeFile()`  [INFERRED]
  cli/internal/plan/sync.go → cli/internal/plan/meta_test.go
- `TestSyncPlanFile_ErrorMissingFile()` --calls--> `SyncPlanFile()`  [INFERRED]
  cli/internal/plan/sync_test.go → cli/internal/plan/sync.go
- `TestParseBacklogFile_NonExistentFile()` --calls--> `ParseBacklogFile()`  [INFERRED]
  cli/internal/plan/backlog_test.go → cli/internal/plan/backlog.go
- `TestScanBacklogDir_EmptyDirectory()` --calls--> `ScanBacklogDir()`  [INFERRED]
  cli/internal/plan/backlog_test.go → cli/internal/plan/backlog.go
- `TestScanBacklogDir_NonExistentDirectory()` --calls--> `ScanBacklogDir()`  [INFERRED]
  cli/internal/plan/backlog_test.go → cli/internal/plan/backlog.go

## Communities

### Community 0 - "Community 0"
Cohesion: 0.09
Nodes (50): formatBacklogTable(), IsOldFormat(), MigrateBacklogDir(), MigrateBacklogFile(), ParseBacklogFile(), ScanBacklog(), ScanBacklogDir(), setupPlanningDir() (+42 more)

### Community 1 - "Community 1"
Cohesion: 0.09
Nodes (43): Logger, Payload, firstInputString(), firstMap(), firstString(), NewLogger(), normalizeToolName(), ParsePayload() (+35 more)

### Community 2 - "Community 2"
Cohesion: 0.09
Nodes (42): Commit(), TestCommit_EmptyMessage_ReturnsError(), ChangedFiles(), CommitsSince(), DiffStat(), FileDiff(), PRInfo(), Run() (+34 more)

### Community 3 - "Community 3"
Cohesion: 0.11
Nodes (35): TestBacklogResult_JSON_ArchiveOmittedWhenEmpty(), TestBacklogResult_JSON_EmptyResult(), TestBacklogResult_JSON_Structure(), FindPlanFile(), GetNextNum(), ParsePlanHeader(), TestFindPlanFile_BranchWithoutSlashStillMatches(), TestFindPlanFile_EmptyDirectory() (+27 more)

### Community 4 - "Community 4"
Cohesion: 0.1
Nodes (32): cacheFilePath(), CheckForUpdate(), fetchLatestVersion(), IsNewer(), parseNum(), readCache(), TestCache_ExpiredAfter24h(), TestCache_InvalidJSON() (+24 more)

### Community 5 - "Community 5"
Cohesion: 0.13
Nodes (15): SessionState, LoadBlueprintConfig(), TestLoadBlueprintConfig_Default(), TestLoadBlueprintConfig_FromFile(), TestRule7ConfigIntegration(), TestSessionState(), NewSessionState(), TestAppendLine_CreatesFileIfNotExists() (+7 more)

### Community 6 - "Community 6"
Cohesion: 0.14
Nodes (16): TestIsAutonomousPipeline_FalseCases(), TestIsAutonomousPipeline_TrueCases(), TestRule10_GitCommitWithoutAISigPasses(), TestRule10_HarmlessBashCommandPasses(), TestRule10_NonBashToolReturnsImmediately(), TestRule10_PhpArtisanMigrateWithoutFreshPasses(), TestRule14_BashWithoutGhPrCreateReturns(), TestRule14_GhPrCreateWithFlowAutoStep5CheckpointPasses() (+8 more)

### Community 7 - "Community 7"
Cohesion: 0.15
Nodes (19): TestAllPaths_BothFilePathAndPatch(), TestAllPaths_NeitherReturnsNil(), TestAllPaths_OnlyFilePath(), TestAllPaths_OnlyPatchPaths(), TestRule1_NonWriteToolReturnsImmediately(), TestRule1_WriteJsxWithSkillAlreadyReadPasses(), TestRule1_WriteToolNonFrontendExtensionPasses(), TestRule6_NonWriteToolReturnsImmediately() (+11 more)

### Community 8 - "Community 8"
Cohesion: 0.3
Nodes (17): SyncResult, SyncPlanFile(), basicPlan(), readFrontmatter(), TestSyncPlanFile_AcceptanceSectionExcluded(), TestSyncPlanFile_AllPhasesComplete(), TestSyncPlanFile_AutoDetect_FindsTodoPlanInBlueprint(), TestSyncPlanFile_BasicTaskCounts() (+9 more)

### Community 9 - "Community 9"
Cohesion: 0.25
Nodes (16): newNilLogger(), TestRule12_BashWithoutGhPrCreateReturns(), TestRule12_GhPrCreateWithNoPlanApprovedPasses(), TestRule12_GhPrCreateWithPlanApprovedAndPlanCheckPasses(), TestRule12_NonBashToolReturns(), TestRule3_TeamCreateRecordsTeamName(), TestRule5_EmptyCommandIgnored(), TestRule5_GhPrCreateInAutonomousPipelineAllowed() (+8 more)

### Community 10 - "Community 10"
Cohesion: 0.18
Nodes (13): BlueprintConfig, TestRule11_BashWithoutGhPrCommentPasses(), TestRule11_GhPrCommentWithFullClaudeReviewPromptPasses(), TestRule11_GhPrCommentWithoutClaudePasses(), TestRule11_NonBashToolReturnsImmediately(), TestRule15_LsBacklogAllowed(), TestRule15_NonBacklogCommandPassesThrough(), TestRule15_NonBashToolIgnored() (+5 more)

### Community 11 - "Community 11"
Cohesion: 0.24
Nodes (13): TestRule2_NestedSkillMdSetsCorrectMarker(), TestRule2_NonReadToolIgnored(), TestRule2_NonSkillReadDoesNotSetMarker(), TestRule2_PlanTemplateMdSetsMarker(), TestRule2_SkillMdReadSetsMarker(), TestRule2_SkillToolFrontendDesignMatchesKnownSkill(), TestRule2_SkillToolInvocationSetsMarker(), TestRule2_TeamExecutionMdSetsMarker() (+5 more)

### Community 12 - "Community 12"
Cohesion: 0.28
Nodes (11): FetchBotComments(), FetchChangedFiles(), FetchDiff(), FetchHumanComments(), FetchInlineComments(), FetchPRChecks(), FetchPRInfo(), FetchReviews() (+3 more)

### Community 13 - "Community 13"
Cohesion: 0.26
Nodes (13): newTestState(), TestRule13_BashWithoutGhPrCreateReturns(), TestRule13_GhPrCreateWithAllAcceptanceCriteriaCheckedPasses(), TestRule13_GhPrCreateWithNoTodoFilePasses(), TestRule13_NonBashToolReturns(), TestRule8_FullSuiteWithParallelAndProcessesAllowed(), TestRule8_HerdCoverageWithFilterAllowed(), TestRule8_NonBashToolIgnored() (+5 more)

### Community 14 - "Community 14"
Cohesion: 0.29
Nodes (7): TestRule3_MultipleStandaloneTasksAccumulateCount(), TestRule3_TaskWithoutTeamNameIncrementsStandaloneCount(), TestRule3_TaskWithTeamNameDoesNotIncrementStandaloneCount(), TestRule3_TeamCreateSetsMarker(), TestRule3_TeamCreateWithoutTeamExecutionReadWarns(), TestRule6_WriteToBlueprintMdWithTemplateAlreadyReadPasses(), rule3TeamCompliance()

### Community 15 - "Community 15"
Cohesion: 0.32
Nodes (6): TestDetectBreakingChange(), TestValidateCommitMessage_AISignatureBlocked(), TestValidateCommitMessage_FailCases(), TestValidateCommitMessage_PassCases(), DetectBreakingChange(), ValidateCommitMessage()

## Knowledge Gaps
- **7 isolated node(s):** `SyncResult`, `BacklogItem`, `BacklogSummary`, `MigrateResult`, `PlanHeader` (+2 more)
  These have ≤1 connection - possible missing edges or undocumented components.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `writeFile()` connect `Community 3` to `Community 0`, `Community 2`, `Community 4`, `Community 5`, `Community 7`, `Community 8`, `Community 13`, `Community 14`?**
  _High betweenness centrality (0.590) - this node is a cross-community bridge._
- **Why does `initRepo()` connect `Community 2` to `Community 3`?**
  _High betweenness centrality (0.200) - this node is a cross-community bridge._
- **Why does `writeBacklogFile()` connect `Community 0` to `Community 3`?**
  _High betweenness centrality (0.161) - this node is a cross-community bridge._
- **Are the 20 inferred relationships involving `writeFile()` (e.g. with `SyncPlanFile()` and `MigrateBacklogFile()`) actually correct?**
  _`writeFile()` has 20 INFERRED edges - model-reasoned connections that need verification._
- **Are the 2 inferred relationships involving `initRepo()` (e.g. with `Run()` and `writeFile()`) actually correct?**
  _`initRepo()` has 2 INFERRED edges - model-reasoned connections that need verification._
- **What connects `SyncResult`, `BacklogItem`, `BacklogSummary` to the rest of the system?**
  _7 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Community 0` be split into smaller, more focused modules?**
  _Cohesion score 0.09 - nodes in this community are weakly interconnected._