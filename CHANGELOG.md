# Change Log

## [0.6.0](https://github.com/fission/fission-workflows/tree/0.6.0) (2018-10-15)
[Full Changelog](https://github.com/fission/fission-workflows/compare/0.5.0...0.6.0)

**Implemented enhancements:**

- Environment Workflow should be general function environment [\#168](https://github.com/fission/fission-workflows/issues/168)
- Support 'output'-field [\#48](https://github.com/fission/fission-workflows/issues/48)
- Implement metric support [\#6](https://github.com/fission/fission-workflows/issues/6)

**Fixed bugs:**

- 0.5.0 appears to break Fission function input [\#172](https://github.com/fission/fission-workflows/issues/172)
- The namespace of workflow environment can not be configured [\#160](https://github.com/fission/fission-workflows/issues/160)
- Fix evaluation queue implementation [\#148](https://github.com/fission/fission-workflows/issues/148)

**Closed issues:**

- Data streams and repository access via Fission? [\#202](https://github.com/fission/fission-workflows/issues/202)
- api docs [\#186](https://github.com/fission/fission-workflows/issues/186)
- Remove wfcli, move workflow functionality to fission CLI [\#68](https://github.com/fission/fission-workflows/issues/68)

**Merged pull requests:**

- Graceful stopping of controller [\#224](https://github.com/fission/fission-workflows/pull/224) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Improved opentracing support [\#222](https://github.com/fission/fission-workflows/pull/222) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Extract environment proxy [\#221](https://github.com/fission/fission-workflows/pull/221) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Bumped Fission dependency to 0.11.0 [\#220](https://github.com/fission/fission-workflows/pull/220) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Added initial makefile [\#219](https://github.com/fission/fission-workflows/pull/219) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Formalize typedvalues implementation [\#218](https://github.com/fission/fission-workflows/pull/218) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add AddTask as an endpoint [\#217](https://github.com/fission/fission-workflows/pull/217) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add LRU and loading caches [\#215](https://github.com/fission/fission-workflows/pull/215) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Bump Fission references to 0.10.0 [\#214](https://github.com/fission/fission-workflows/pull/214) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Minor CLI improvements [\#213](https://github.com/fission/fission-workflows/pull/213) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Force k8s object removal in e2e cleanup [\#212](https://github.com/fission/fission-workflows/pull/212) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Generate event type identifiers [\#211](https://github.com/fission/fission-workflows/pull/211) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Added invocation and workflow stores for typed entity retrieval [\#210](https://github.com/fission/fission-workflows/pull/210) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Constrain memory usage in the in-memory backend [\#209](https://github.com/fission/fission-workflows/pull/209) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Revert "improve mechanism inside invocation rule" [\#208](https://github.com/fission/fission-workflows/pull/208) ([erwinvaneyk](https://github.com/erwinvaneyk))
- HTTP Runtime [\#207](https://github.com/fission/fission-workflows/pull/207) ([erwinvaneyk](https://github.com/erwinvaneyk))
- improve mechanism inside invocation rule [\#206](https://github.com/fission/fission-workflows/pull/206) ([xiekeyang](https://github.com/xiekeyang))
- Remove RuleHasCompleted structure [\#199](https://github.com/fission/fission-workflows/pull/199) ([xiekeyang](https://github.com/xiekeyang))
- Support output transformations [\#194](https://github.com/fission/fission-workflows/pull/194) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add support for jsonpb-encoded workflow specs [\#193](https://github.com/fission/fission-workflows/pull/193) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add workqueue to Invocation Controller [\#192](https://github.com/fission/fission-workflows/pull/192) ([xiekeyang](https://github.com/xiekeyang))
- Opentracing Support [\#185](https://github.com/fission/fission-workflows/pull/185) ([erwinvaneyk](https://github.com/erwinvaneyk))
- close response body [\#182](https://github.com/fission/fission-workflows/pull/182) ([xiekeyang](https://github.com/xiekeyang))
- Fix HTTP handler of health checking [\#181](https://github.com/fission/fission-workflows/pull/181) ([xiekeyang](https://github.com/xiekeyang))
- Event structuring [\#179](https://github.com/fission/fission-workflows/pull/179) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Propagate HTTP invocation context [\#177](https://github.com/fission/fission-workflows/pull/177) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Parse and Resolve the Namespace of Fission function [\#176](https://github.com/fission/fission-workflows/pull/176) ([xiekeyang](https://github.com/xiekeyang))
- Added environment as a kubernetes resource definition [\#175](https://github.com/fission/fission-workflows/pull/175) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update compiling instructions and rename potential conflicting wfcli directory [\#171](https://github.com/fission/fission-workflows/pull/171) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Parallelize task executions [\#170](https://github.com/fission/fission-workflows/pull/170) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Use priority queue instead of queue [\#167](https://github.com/fission/fission-workflows/pull/167) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Updated Fission dependency to 0.9.1 [\#166](https://github.com/fission/fission-workflows/pull/166) ([erwinvaneyk](https://github.com/erwinvaneyk))
- add configurable runtime parameters to chart [\#165](https://github.com/fission/fission-workflows/pull/165) ([xiekeyang](https://github.com/xiekeyang))
- Make CLI compatible with Fission CLI plugin interface [\#158](https://github.com/fission/fission-workflows/pull/158) ([erwinvaneyk](https://github.com/erwinvaneyk))

## [0.5.0](https://github.com/fission/fission-workflows/tree/0.5.0) (2018-07-11)
[Full Changelog](https://github.com/fission/fission-workflows/compare/0.4.0...0.5.0)

**Implemented enhancements:**

- Attach invocation context to logs [\#86](https://github.com/fission/fission-workflows/issues/86)

**Closed issues:**

- Change Workflows to talk to router [\#84](https://github.com/fission/fission-workflows/issues/84)

**Merged pull requests:**

- 0.4.0 -\> 0.5.0 [\#164](https://github.com/fission/fission-workflows/pull/164) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update commands to setup fission functions [\#163](https://github.com/fission/fission-workflows/pull/163) ([beevelop](https://github.com/beevelop))
- Update wfcli docs [\#162](https://github.com/fission/fission-workflows/pull/162) ([beevelop](https://github.com/beevelop))
- add NOBUILD ARG to script [\#161](https://github.com/fission/fission-workflows/pull/161) ([xiekeyang](https://github.com/xiekeyang))
- YAML API improvements [\#159](https://github.com/fission/fission-workflows/pull/159) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Log correlation [\#157](https://github.com/fission/fission-workflows/pull/157) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Set content-length when setting body [\#156](https://github.com/fission/fission-workflows/pull/156) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Cleanup labels package [\#154](https://github.com/fission/fission-workflows/pull/154) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Namespace proto files [\#153](https://github.com/fission/fission-workflows/pull/153) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Listen to system termination signals [\#152](https://github.com/fission/fission-workflows/pull/152) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Match validation errors to IllegalArgument HTTP statuses [\#151](https://github.com/fission/fission-workflows/pull/151) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Fission integration tests [\#121](https://github.com/fission/fission-workflows/pull/121) ([erwinvaneyk](https://github.com/erwinvaneyk))

## [0.4.0](https://github.com/fission/fission-workflows/tree/0.4.0) (2018-06-07)
[Full Changelog](https://github.com/fission/fission-workflows/compare/0.3.0...0.4.0)

**Merged pull requests:**

- 0.3.0 -\> 0.4.0 [\#149](https://github.com/fission/fission-workflows/pull/149) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Merge API packages [\#147](https://github.com/fission/fission-workflows/pull/147) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Include git info in versioning [\#146](https://github.com/fission/fission-workflows/pull/146) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Prometheus integration [\#122](https://github.com/fission/fission-workflows/pull/122) ([erwinvaneyk](https://github.com/erwinvaneyk))

## [0.3.0](https://github.com/fission/fission-workflows/tree/0.3.0) (2018-05-17)
[Full Changelog](https://github.com/fission/fission-workflows/compare/0.2.0...0.3.0)

**Implemented enhancements:**

- Add API change warning to documentation [\#95](https://github.com/fission/fission-workflows/issues/95)
- Add cancel command to wfcli [\#88](https://github.com/fission/fission-workflows/issues/88)
- Validation check for circular dependencies [\#42](https://github.com/fission/fission-workflows/issues/42)
- Support headers \(and other metadata\) as input to workflow [\#96](https://github.com/fission/fission-workflows/issues/96)
- Support inline workflows [\#44](https://github.com/fission/fission-workflows/issues/44)
- Support passing query and headers to fission functions [\#37](https://github.com/fission/fission-workflows/issues/37)

**Fixed bugs:**

- Fix builder errors when supplying an archive [\#139](https://github.com/fission/fission-workflows/issues/139)
- Release latest fission workflow version [\#138](https://github.com/fission/fission-workflows/issues/138)
- Limit number of parallel subscribers to backing event store [\#85](https://github.com/fission/fission-workflows/issues/85)
- duplicate task invocations [\#77](https://github.com/fission/fission-workflows/issues/77)
- Infinite fail loop : limit function rate / retries [\#71](https://github.com/fission/fission-workflows/issues/71)
- Workflows install fails if fission is not installed in the "fission" namespace [\#69](https://github.com/fission/fission-workflows/issues/69)

**Closed issues:**

- Example fortunewhale workflow failing [\#141](https://github.com/fission/fission-workflows/issues/141)
- Fission workflow did not work correctly [\#137](https://github.com/fission/fission-workflows/issues/137)
- Formalize function call API [\#136](https://github.com/fission/fission-workflows/issues/136)
- Errors when installing wfcli [\#133](https://github.com/fission/fission-workflows/issues/133)
- `fission function logs` should show meaningful logs for workflows [\#125](https://github.com/fission/fission-workflows/issues/125)
- Install instructions are missing wfcli [\#124](https://github.com/fission/fission-workflows/issues/124)
- Document functionality of the query parser [\#43](https://github.com/fission/fission-workflows/issues/43)
- Add /healthz [\#123](https://github.com/fission/fission-workflows/issues/123)
- Add Fission e2e test  [\#40](https://github.com/fission/fission-workflows/issues/40)

**Merged pull requests:**

- Fission Workflows 0.3.0 [\#145](https://github.com/fission/fission-workflows/pull/145) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add demo kubecon 2018 [\#142](https://github.com/fission/fission-workflows/pull/142) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add chmod line to the installation instructions [\#134](https://github.com/fission/fission-workflows/pull/134) ([erwinvaneyk](https://github.com/erwinvaneyk))
- fission 0.6.0 -\> 0.6.1 [\#132](https://github.com/fission/fission-workflows/pull/132) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Remove swagger golang client from the wfcli tool [\#131](https://github.com/fission/fission-workflows/pull/131) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add installation instructions for wfcli client [\#130](https://github.com/fission/fission-workflows/pull/130) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Bump controller tick speed to 100 ms [\#128](https://github.com/fission/fission-workflows/pull/128) ([erwinvaneyk](https://github.com/erwinvaneyk))
- wfcli: cancel, invoke, halt, resume [\#127](https://github.com/fission/fission-workflows/pull/127) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add readiness and liveniness probes [\#126](https://github.com/fission/fission-workflows/pull/126) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Formalize evaluation state in controllers [\#119](https://github.com/fission/fission-workflows/pull/119) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Improve workflow definition [\#118](https://github.com/fission/fission-workflows/pull/118) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Centralize the fragemented integration tests and utils  [\#117](https://github.com/fission/fission-workflows/pull/117) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update glide dependencies [\#116](https://github.com/fission/fission-workflows/pull/116) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add invocationId to the TaskInvocation to avoid dangling tasks [\#115](https://github.com/fission/fission-workflows/pull/115) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Extract resolver and parser implementations from workflow api [\#114](https://github.com/fission/fission-workflows/pull/114) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Improve \(docker\) build scripts [\#113](https://github.com/fission/fission-workflows/pull/113) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Support for dynamic workflows [\#112](https://github.com/fission/fission-workflows/pull/112) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add improved graph and validation support [\#111](https://github.com/fission/fission-workflows/pull/111) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add Javascript and Repeat functions to internal fnenv [\#110](https://github.com/fission/fission-workflows/pull/110) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Improve concurrency of the event store implementation [\#109](https://github.com/fission/fission-workflows/pull/109) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add prewarm functionality to Fission fnenv [\#108](https://github.com/fission/fission-workflows/pull/108) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add input expressions documentation [\#107](https://github.com/fission/fission-workflows/pull/107) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add fnenv.Notifier interface + restructuring of fnenv package [\#106](https://github.com/fission/fission-workflows/pull/106) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Break up CLI into multiple files [\#105](https://github.com/fission/fission-workflows/pull/105) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update Travis to use a new GKE cluster for e2e tests [\#104](https://github.com/fission/fission-workflows/pull/104) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Stability fixes NATS event store implementation [\#103](https://github.com/fission/fission-workflows/pull/103) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Integrate e2e with Travis/GKE [\#102](https://github.com/fission/fission-workflows/pull/102) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Improve logging and concurrency in controller [\#93](https://github.com/fission/fission-workflows/pull/93) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Introduce e2e tests [\#92](https://github.com/fission/fission-workflows/pull/92) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Fix bash variable errors at workflow builder cmd [\#140](https://github.com/fission/fission-workflows/pull/140) ([thenamly](https://github.com/thenamly))
- Add control flow / utility functions to the workflow engine  [\#135](https://github.com/fission/fission-workflows/pull/135) ([erwinvaneyk](https://github.com/erwinvaneyk))

## [0.2.0](https://github.com/fission/fission-workflows/tree/0.2.0) (2018-01-22)
[Full Changelog](https://github.com/fission/fission-workflows/compare/0.1.1...0.2.0)

**Fixed bugs:**

- Improve expression type parsing [\#49](https://github.com/fission/fission-workflows/issues/49)

**Closed issues:**

- Helm chart notes: incorrect image tag in command [\#72](https://github.com/fission/fission-workflows/issues/72)
- Do we need Bazel? [\#70](https://github.com/fission/fission-workflows/issues/70)
- Automate yaml -\> json wf definition compilation [\#60](https://github.com/fission/fission-workflows/issues/60)
- Investigate poolmgr cleanup of workflow engine [\#50](https://github.com/fission/fission-workflows/issues/50)
- Update README examples to be functional [\#41](https://github.com/fission/fission-workflows/issues/41)
- Remove TypedValue serialization from json format [\#39](https://github.com/fission/fission-workflows/issues/39)
- Create separate Api-only service  [\#30](https://github.com/fission/fission-workflows/issues/30)
- Create Helm chart [\#29](https://github.com/fission/fission-workflows/issues/29)
- Use Fission Builder to parse yaml -\> json [\#28](https://github.com/fission/fission-workflows/issues/28)
- Workflow CLI [\#8](https://github.com/fission/fission-workflows/issues/8)
- Workflow-engine should be just another Fission environment [\#5](https://github.com/fission/fission-workflows/issues/5)

**Merged pull requests:**

- Release 0.2.0 [\#100](https://github.com/fission/fission-workflows/pull/100) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add figures to readme [\#98](https://github.com/fission/fission-workflows/pull/98) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add support for headers and query params to workflow invocations [\#97](https://github.com/fission/fission-workflows/pull/97) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add go vet check to project [\#94](https://github.com/fission/fission-workflows/pull/94) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add support for inputs specifying task metadata [\#91](https://github.com/fission/fission-workflows/pull/91) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add environment variables to deployment [\#90](https://github.com/fission/fission-workflows/pull/90) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update Fission dependency [\#83](https://github.com/fission/fission-workflows/pull/83) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Extract tag from helm deployment [\#82](https://github.com/fission/fission-workflows/pull/82) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Remove bazel build files [\#81](https://github.com/fission/fission-workflows/pull/81) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add ReadItLater example [\#80](https://github.com/fission/fission-workflows/pull/80) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Use Fission Builders for the yaml parsing [\#79](https://github.com/fission/fission-workflows/pull/79) ([erwinvaneyk](https://github.com/erwinvaneyk))

## [0.1.1](https://github.com/fission/fission-workflows/tree/0.1.1) (2017-10-01)
**Closed issues:**

- Make internal functions pluggable [\#64](https://github.com/fission/fission-workflows/issues/64)
- Fixing naming inconsistencies [\#31](https://github.com/fission/fission-workflows/issues/31)
- Add documentation [\#7](https://github.com/fission/fission-workflows/issues/7)
- TODOs [\#3](https://github.com/fission/fission-workflows/issues/3)

**Merged pull requests:**

- 0.1.1 [\#63](https://github.com/fission/fission-workflows/pull/63) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update README and Install [\#62](https://github.com/fission/fission-workflows/pull/62) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update fission-workflows chart [\#61](https://github.com/fission/fission-workflows/pull/61) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Stability improvements [\#59](https://github.com/fission/fission-workflows/pull/59) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Change expression parsing to expect '{ ... }' to delimit expressions [\#57](https://github.com/fission/fission-workflows/pull/57) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add Slackweather example [\#56](https://github.com/fission/fission-workflows/pull/56) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Object input support [\#55](https://github.com/fission/fission-workflows/pull/55) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Move deployment to a helm chart [\#53](https://github.com/fission/fission-workflows/pull/53) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add compose function [\#52](https://github.com/fission/fission-workflows/pull/52) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Fission Integration part 2 [\#45](https://github.com/fission/fission-workflows/pull/45) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Integrate workflow with Fission as an environment  [\#36](https://github.com/fission/fission-workflows/pull/36) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Initial Fission Environment integration [\#35](https://github.com/fission/fission-workflows/pull/35) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add fission.FunctionRef parser/formatter [\#33](https://github.com/fission/fission-workflows/pull/33) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Fixed naming inconsistencies [\#32](https://github.com/fission/fission-workflows/pull/32) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add internal sleep function [\#26](https://github.com/fission/fission-workflows/pull/26) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add functionality for modifying the control flow  [\#24](https://github.com/fission/fission-workflows/pull/24) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Timeout + scope isolation in Otto interpreter [\#23](https://github.com/fission/fission-workflows/pull/23) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add full transformer and selector support [\#22](https://github.com/fission/fission-workflows/pull/22) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Improve controller notifications [\#21](https://github.com/fission/fission-workflows/pull/21) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add slack notification token to travis config [\#20](https://github.com/fission/fission-workflows/pull/20) ([soamvasani](https://github.com/soamvasani))
- Add Travis CI [\#19](https://github.com/fission/fission-workflows/pull/19) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Roadmap [\#18](https://github.com/fission/fission-workflows/pull/18) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add documentation [\#17](https://github.com/fission/fission-workflows/pull/17) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Introduce native runtime and selectors [\#16](https://github.com/fission/fission-workflows/pull/16) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Update input mapping [\#15](https://github.com/fission/fission-workflows/pull/15) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Add License \(Apache v2\) and Code Of Conduct \(Contributor Covenant 1.4\) [\#14](https://github.com/fission/fission-workflows/pull/14) ([soamvasani](https://github.com/soamvasani))
- Initial documentation [\#12](https://github.com/fission/fission-workflows/pull/12) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Initial prototype [\#2](https://github.com/fission/fission-workflows/pull/2) ([erwinvaneyk](https://github.com/erwinvaneyk))
- Build and scaffolding setup [\#1](https://github.com/fission/fission-workflows/pull/1) ([erwinvaneyk](https://github.com/erwinvaneyk))



\* *This Change Log was automatically generated by [github_changelog_generator](https://github.com/skywinder/Github-Changelog-Generator)*