/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
(function(global, factory) { /* global define, require, module */

    /* AMD */ if (typeof define === 'function' && define.amd)
        define(["protobufjs/minimal"], factory);

    /* CommonJS */ else if (typeof require === 'function' && typeof module === 'object' && module && module.exports)
        module.exports = factory(require("protobufjs/minimal"));

})(this, function($protobuf) {
    "use strict";

    // Common aliases
    var $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;
    
    // Exported root namespace
    var $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});
    
    $root.fission = (function() {
    
        /**
         * Namespace fission.
         * @exports fission
         * @namespace
         */
        var fission = {};
    
        fission.workflows = (function() {
    
            /**
             * Namespace workflows.
             * @memberof fission
             * @namespace
             */
            var workflows = {};
    
            workflows.types = (function() {
    
                /**
                 * Namespace types.
                 * @memberof fission.workflows
                 * @namespace
                 */
                var types = {};
    
                types.Workflow = (function() {
    
                    /**
                     * Properties of a Workflow.
                     * @memberof fission.workflows.types
                     * @interface IWorkflow
                     * @property {fission.workflows.types.IObjectMetadata|null} [metadata] Workflow metadata
                     * @property {fission.workflows.types.IWorkflowSpec|null} [spec] Workflow spec
                     * @property {fission.workflows.types.IWorkflowStatus|null} [status] Workflow status
                     */
    
                    /**
                     * Constructs a new Workflow.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a Workflow.
                     * @implements IWorkflow
                     * @constructor
                     * @param {fission.workflows.types.IWorkflow=} [properties] Properties to set
                     */
                    function Workflow(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * Workflow metadata.
                     * @member {fission.workflows.types.IObjectMetadata|null|undefined} metadata
                     * @memberof fission.workflows.types.Workflow
                     * @instance
                     */
                    Workflow.prototype.metadata = null;
    
                    /**
                     * Workflow spec.
                     * @member {fission.workflows.types.IWorkflowSpec|null|undefined} spec
                     * @memberof fission.workflows.types.Workflow
                     * @instance
                     */
                    Workflow.prototype.spec = null;
    
                    /**
                     * Workflow status.
                     * @member {fission.workflows.types.IWorkflowStatus|null|undefined} status
                     * @memberof fission.workflows.types.Workflow
                     * @instance
                     */
                    Workflow.prototype.status = null;
    
                    /**
                     * Creates a new Workflow instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {fission.workflows.types.IWorkflow=} [properties] Properties to set
                     * @returns {fission.workflows.types.Workflow} Workflow instance
                     */
                    Workflow.create = function create(properties) {
                        return new Workflow(properties);
                    };
    
                    /**
                     * Encodes the specified Workflow message. Does not implicitly {@link fission.workflows.types.Workflow.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {fission.workflows.types.IWorkflow} message Workflow message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    Workflow.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            $root.fission.workflows.types.ObjectMetadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            $root.fission.workflows.types.WorkflowSpec.encode(message.spec, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.status != null && message.hasOwnProperty("status"))
                            $root.fission.workflows.types.WorkflowStatus.encode(message.status, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified Workflow message, length delimited. Does not implicitly {@link fission.workflows.types.Workflow.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {fission.workflows.types.IWorkflow} message Workflow message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    Workflow.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a Workflow message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.Workflow} Workflow
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    Workflow.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.Workflow();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.metadata = $root.fission.workflows.types.ObjectMetadata.decode(reader, reader.uint32());
                                break;
                            case 2:
                                message.spec = $root.fission.workflows.types.WorkflowSpec.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.status = $root.fission.workflows.types.WorkflowStatus.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a Workflow message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.Workflow} Workflow
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    Workflow.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a Workflow message.
                     * @function verify
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    Workflow.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.metadata != null && message.hasOwnProperty("metadata")) {
                            var error = $root.fission.workflows.types.ObjectMetadata.verify(message.metadata);
                            if (error)
                                return "metadata." + error;
                        }
                        if (message.spec != null && message.hasOwnProperty("spec")) {
                            var error = $root.fission.workflows.types.WorkflowSpec.verify(message.spec);
                            if (error)
                                return "spec." + error;
                        }
                        if (message.status != null && message.hasOwnProperty("status")) {
                            var error = $root.fission.workflows.types.WorkflowStatus.verify(message.status);
                            if (error)
                                return "status." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a Workflow message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.Workflow} Workflow
                     */
                    Workflow.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.Workflow)
                            return object;
                        var message = new $root.fission.workflows.types.Workflow();
                        if (object.metadata != null) {
                            if (typeof object.metadata !== "object")
                                throw TypeError(".fission.workflows.types.Workflow.metadata: object expected");
                            message.metadata = $root.fission.workflows.types.ObjectMetadata.fromObject(object.metadata);
                        }
                        if (object.spec != null) {
                            if (typeof object.spec !== "object")
                                throw TypeError(".fission.workflows.types.Workflow.spec: object expected");
                            message.spec = $root.fission.workflows.types.WorkflowSpec.fromObject(object.spec);
                        }
                        if (object.status != null) {
                            if (typeof object.status !== "object")
                                throw TypeError(".fission.workflows.types.Workflow.status: object expected");
                            message.status = $root.fission.workflows.types.WorkflowStatus.fromObject(object.status);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a Workflow message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.Workflow
                     * @static
                     * @param {fission.workflows.types.Workflow} message Workflow
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    Workflow.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.metadata = null;
                            object.spec = null;
                            object.status = null;
                        }
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            object.metadata = $root.fission.workflows.types.ObjectMetadata.toObject(message.metadata, options);
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            object.spec = $root.fission.workflows.types.WorkflowSpec.toObject(message.spec, options);
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = $root.fission.workflows.types.WorkflowStatus.toObject(message.status, options);
                        return object;
                    };
    
                    /**
                     * Converts this Workflow to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.Workflow
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    Workflow.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return Workflow;
                })();
    
                types.WorkflowSpec = (function() {
    
                    /**
                     * Properties of a WorkflowSpec.
                     * @memberof fission.workflows.types
                     * @interface IWorkflowSpec
                     * @property {string|null} [apiVersion] WorkflowSpec apiVersion
                     * @property {Object.<string,fission.workflows.types.ITaskSpec>|null} [tasks] WorkflowSpec tasks
                     * @property {string|null} [outputTask] WorkflowSpec outputTask
                     * @property {string|null} [description] WorkflowSpec description
                     * @property {string|null} [forceId] WorkflowSpec forceId
                     * @property {string|null} [name] WorkflowSpec name
                     * @property {boolean|null} [internal] WorkflowSpec internal
                     */
    
                    /**
                     * Constructs a new WorkflowSpec.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a WorkflowSpec.
                     * @implements IWorkflowSpec
                     * @constructor
                     * @param {fission.workflows.types.IWorkflowSpec=} [properties] Properties to set
                     */
                    function WorkflowSpec(properties) {
                        this.tasks = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * WorkflowSpec apiVersion.
                     * @member {string} apiVersion
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     */
                    WorkflowSpec.prototype.apiVersion = "";
    
                    /**
                     * WorkflowSpec tasks.
                     * @member {Object.<string,fission.workflows.types.ITaskSpec>} tasks
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     */
                    WorkflowSpec.prototype.tasks = $util.emptyObject;
    
                    /**
                     * WorkflowSpec outputTask.
                     * @member {string} outputTask
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     */
                    WorkflowSpec.prototype.outputTask = "";
    
                    /**
                     * WorkflowSpec description.
                     * @member {string} description
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     */
                    WorkflowSpec.prototype.description = "";
    
                    /**
                     * WorkflowSpec forceId.
                     * @member {string} forceId
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     */
                    WorkflowSpec.prototype.forceId = "";
    
                    /**
                     * WorkflowSpec name.
                     * @member {string} name
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     */
                    WorkflowSpec.prototype.name = "";
    
                    /**
                     * WorkflowSpec internal.
                     * @member {boolean} internal
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     */
                    WorkflowSpec.prototype.internal = false;
    
                    /**
                     * Creates a new WorkflowSpec instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {fission.workflows.types.IWorkflowSpec=} [properties] Properties to set
                     * @returns {fission.workflows.types.WorkflowSpec} WorkflowSpec instance
                     */
                    WorkflowSpec.create = function create(properties) {
                        return new WorkflowSpec(properties);
                    };
    
                    /**
                     * Encodes the specified WorkflowSpec message. Does not implicitly {@link fission.workflows.types.WorkflowSpec.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {fission.workflows.types.IWorkflowSpec} message WorkflowSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowSpec.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.apiVersion != null && message.hasOwnProperty("apiVersion"))
                            writer.uint32(/* id 1, wireType 2 =*/10).string(message.apiVersion);
                        if (message.tasks != null && message.hasOwnProperty("tasks"))
                            for (var keys = Object.keys(message.tasks), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 2, wireType 2 =*/18).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TaskSpec.encode(message.tasks[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.outputTask != null && message.hasOwnProperty("outputTask"))
                            writer.uint32(/* id 3, wireType 2 =*/26).string(message.outputTask);
                        if (message.description != null && message.hasOwnProperty("description"))
                            writer.uint32(/* id 4, wireType 2 =*/34).string(message.description);
                        if (message.forceId != null && message.hasOwnProperty("forceId"))
                            writer.uint32(/* id 5, wireType 2 =*/42).string(message.forceId);
                        if (message.name != null && message.hasOwnProperty("name"))
                            writer.uint32(/* id 6, wireType 2 =*/50).string(message.name);
                        if (message.internal != null && message.hasOwnProperty("internal"))
                            writer.uint32(/* id 7, wireType 0 =*/56).bool(message.internal);
                        return writer;
                    };
    
                    /**
                     * Encodes the specified WorkflowSpec message, length delimited. Does not implicitly {@link fission.workflows.types.WorkflowSpec.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {fission.workflows.types.IWorkflowSpec} message WorkflowSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowSpec.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a WorkflowSpec message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.WorkflowSpec} WorkflowSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowSpec.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.WorkflowSpec(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.apiVersion = reader.string();
                                break;
                            case 2:
                                reader.skip().pos++;
                                if (message.tasks === $util.emptyObject)
                                    message.tasks = {};
                                key = reader.string();
                                reader.pos++;
                                message.tasks[key] = $root.fission.workflows.types.TaskSpec.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.outputTask = reader.string();
                                break;
                            case 4:
                                message.description = reader.string();
                                break;
                            case 5:
                                message.forceId = reader.string();
                                break;
                            case 6:
                                message.name = reader.string();
                                break;
                            case 7:
                                message.internal = reader.bool();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a WorkflowSpec message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.WorkflowSpec} WorkflowSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowSpec.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a WorkflowSpec message.
                     * @function verify
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    WorkflowSpec.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.apiVersion != null && message.hasOwnProperty("apiVersion"))
                            if (!$util.isString(message.apiVersion))
                                return "apiVersion: string expected";
                        if (message.tasks != null && message.hasOwnProperty("tasks")) {
                            if (!$util.isObject(message.tasks))
                                return "tasks: object expected";
                            var key = Object.keys(message.tasks);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TaskSpec.verify(message.tasks[key[i]]);
                                if (error)
                                    return "tasks." + error;
                            }
                        }
                        if (message.outputTask != null && message.hasOwnProperty("outputTask"))
                            if (!$util.isString(message.outputTask))
                                return "outputTask: string expected";
                        if (message.description != null && message.hasOwnProperty("description"))
                            if (!$util.isString(message.description))
                                return "description: string expected";
                        if (message.forceId != null && message.hasOwnProperty("forceId"))
                            if (!$util.isString(message.forceId))
                                return "forceId: string expected";
                        if (message.name != null && message.hasOwnProperty("name"))
                            if (!$util.isString(message.name))
                                return "name: string expected";
                        if (message.internal != null && message.hasOwnProperty("internal"))
                            if (typeof message.internal !== "boolean")
                                return "internal: boolean expected";
                        return null;
                    };
    
                    /**
                     * Creates a WorkflowSpec message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.WorkflowSpec} WorkflowSpec
                     */
                    WorkflowSpec.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.WorkflowSpec)
                            return object;
                        var message = new $root.fission.workflows.types.WorkflowSpec();
                        if (object.apiVersion != null)
                            message.apiVersion = String(object.apiVersion);
                        if (object.tasks) {
                            if (typeof object.tasks !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowSpec.tasks: object expected");
                            message.tasks = {};
                            for (var keys = Object.keys(object.tasks), i = 0; i < keys.length; ++i) {
                                if (typeof object.tasks[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.WorkflowSpec.tasks: object expected");
                                message.tasks[keys[i]] = $root.fission.workflows.types.TaskSpec.fromObject(object.tasks[keys[i]]);
                            }
                        }
                        if (object.outputTask != null)
                            message.outputTask = String(object.outputTask);
                        if (object.description != null)
                            message.description = String(object.description);
                        if (object.forceId != null)
                            message.forceId = String(object.forceId);
                        if (object.name != null)
                            message.name = String(object.name);
                        if (object.internal != null)
                            message.internal = Boolean(object.internal);
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a WorkflowSpec message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @static
                     * @param {fission.workflows.types.WorkflowSpec} message WorkflowSpec
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    WorkflowSpec.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults)
                            object.tasks = {};
                        if (options.defaults) {
                            object.apiVersion = "";
                            object.outputTask = "";
                            object.description = "";
                            object.forceId = "";
                            object.name = "";
                            object.internal = false;
                        }
                        if (message.apiVersion != null && message.hasOwnProperty("apiVersion"))
                            object.apiVersion = message.apiVersion;
                        var keys2;
                        if (message.tasks && (keys2 = Object.keys(message.tasks)).length) {
                            object.tasks = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.tasks[keys2[j]] = $root.fission.workflows.types.TaskSpec.toObject(message.tasks[keys2[j]], options);
                        }
                        if (message.outputTask != null && message.hasOwnProperty("outputTask"))
                            object.outputTask = message.outputTask;
                        if (message.description != null && message.hasOwnProperty("description"))
                            object.description = message.description;
                        if (message.forceId != null && message.hasOwnProperty("forceId"))
                            object.forceId = message.forceId;
                        if (message.name != null && message.hasOwnProperty("name"))
                            object.name = message.name;
                        if (message.internal != null && message.hasOwnProperty("internal"))
                            object.internal = message.internal;
                        return object;
                    };
    
                    /**
                     * Converts this WorkflowSpec to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.WorkflowSpec
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    WorkflowSpec.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return WorkflowSpec;
                })();
    
                types.WorkflowStatus = (function() {
    
                    /**
                     * Properties of a WorkflowStatus.
                     * @memberof fission.workflows.types
                     * @interface IWorkflowStatus
                     * @property {fission.workflows.types.WorkflowStatus.Status|null} [status] WorkflowStatus status
                     * @property {google.protobuf.ITimestamp|null} [updatedAt] WorkflowStatus updatedAt
                     * @property {Object.<string,fission.workflows.types.ITaskStatus>|null} [tasks] WorkflowStatus tasks
                     * @property {fission.workflows.types.IError|null} [error] WorkflowStatus error
                     */
    
                    /**
                     * Constructs a new WorkflowStatus.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a WorkflowStatus.
                     * @implements IWorkflowStatus
                     * @constructor
                     * @param {fission.workflows.types.IWorkflowStatus=} [properties] Properties to set
                     */
                    function WorkflowStatus(properties) {
                        this.tasks = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * WorkflowStatus status.
                     * @member {fission.workflows.types.WorkflowStatus.Status} status
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @instance
                     */
                    WorkflowStatus.prototype.status = 0;
    
                    /**
                     * WorkflowStatus updatedAt.
                     * @member {google.protobuf.ITimestamp|null|undefined} updatedAt
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @instance
                     */
                    WorkflowStatus.prototype.updatedAt = null;
    
                    /**
                     * WorkflowStatus tasks.
                     * @member {Object.<string,fission.workflows.types.ITaskStatus>} tasks
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @instance
                     */
                    WorkflowStatus.prototype.tasks = $util.emptyObject;
    
                    /**
                     * WorkflowStatus error.
                     * @member {fission.workflows.types.IError|null|undefined} error
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @instance
                     */
                    WorkflowStatus.prototype.error = null;
    
                    /**
                     * Creates a new WorkflowStatus instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {fission.workflows.types.IWorkflowStatus=} [properties] Properties to set
                     * @returns {fission.workflows.types.WorkflowStatus} WorkflowStatus instance
                     */
                    WorkflowStatus.create = function create(properties) {
                        return new WorkflowStatus(properties);
                    };
    
                    /**
                     * Encodes the specified WorkflowStatus message. Does not implicitly {@link fission.workflows.types.WorkflowStatus.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {fission.workflows.types.IWorkflowStatus} message WorkflowStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowStatus.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.status != null && message.hasOwnProperty("status"))
                            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.status);
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            $root.google.protobuf.Timestamp.encode(message.updatedAt, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.tasks != null && message.hasOwnProperty("tasks"))
                            for (var keys = Object.keys(message.tasks), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 3, wireType 2 =*/26).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TaskStatus.encode(message.tasks[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.error != null && message.hasOwnProperty("error"))
                            $root.fission.workflows.types.Error.encode(message.error, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified WorkflowStatus message, length delimited. Does not implicitly {@link fission.workflows.types.WorkflowStatus.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {fission.workflows.types.IWorkflowStatus} message WorkflowStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowStatus.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a WorkflowStatus message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.WorkflowStatus} WorkflowStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowStatus.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.WorkflowStatus(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.status = reader.int32();
                                break;
                            case 2:
                                message.updatedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                                break;
                            case 3:
                                reader.skip().pos++;
                                if (message.tasks === $util.emptyObject)
                                    message.tasks = {};
                                key = reader.string();
                                reader.pos++;
                                message.tasks[key] = $root.fission.workflows.types.TaskStatus.decode(reader, reader.uint32());
                                break;
                            case 4:
                                message.error = $root.fission.workflows.types.Error.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a WorkflowStatus message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.WorkflowStatus} WorkflowStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowStatus.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a WorkflowStatus message.
                     * @function verify
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    WorkflowStatus.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.status != null && message.hasOwnProperty("status"))
                            switch (message.status) {
                            default:
                                return "status: enum value expected";
                            case 0:
                            case 2:
                            case 3:
                            case 4:
                                break;
                            }
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt")) {
                            var error = $root.google.protobuf.Timestamp.verify(message.updatedAt);
                            if (error)
                                return "updatedAt." + error;
                        }
                        if (message.tasks != null && message.hasOwnProperty("tasks")) {
                            if (!$util.isObject(message.tasks))
                                return "tasks: object expected";
                            var key = Object.keys(message.tasks);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TaskStatus.verify(message.tasks[key[i]]);
                                if (error)
                                    return "tasks." + error;
                            }
                        }
                        if (message.error != null && message.hasOwnProperty("error")) {
                            var error = $root.fission.workflows.types.Error.verify(message.error);
                            if (error)
                                return "error." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a WorkflowStatus message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.WorkflowStatus} WorkflowStatus
                     */
                    WorkflowStatus.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.WorkflowStatus)
                            return object;
                        var message = new $root.fission.workflows.types.WorkflowStatus();
                        switch (object.status) {
                        case "PENDING":
                        case 0:
                            message.status = 0;
                            break;
                        case "READY":
                        case 2:
                            message.status = 2;
                            break;
                        case "FAILED":
                        case 3:
                            message.status = 3;
                            break;
                        case "DELETED":
                        case 4:
                            message.status = 4;
                            break;
                        }
                        if (object.updatedAt != null) {
                            if (typeof object.updatedAt !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowStatus.updatedAt: object expected");
                            message.updatedAt = $root.google.protobuf.Timestamp.fromObject(object.updatedAt);
                        }
                        if (object.tasks) {
                            if (typeof object.tasks !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowStatus.tasks: object expected");
                            message.tasks = {};
                            for (var keys = Object.keys(object.tasks), i = 0; i < keys.length; ++i) {
                                if (typeof object.tasks[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.WorkflowStatus.tasks: object expected");
                                message.tasks[keys[i]] = $root.fission.workflows.types.TaskStatus.fromObject(object.tasks[keys[i]]);
                            }
                        }
                        if (object.error != null) {
                            if (typeof object.error !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowStatus.error: object expected");
                            message.error = $root.fission.workflows.types.Error.fromObject(object.error);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a WorkflowStatus message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @static
                     * @param {fission.workflows.types.WorkflowStatus} message WorkflowStatus
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    WorkflowStatus.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults)
                            object.tasks = {};
                        if (options.defaults) {
                            object.status = options.enums === String ? "PENDING" : 0;
                            object.updatedAt = null;
                            object.error = null;
                        }
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = options.enums === String ? $root.fission.workflows.types.WorkflowStatus.Status[message.status] : message.status;
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            object.updatedAt = $root.google.protobuf.Timestamp.toObject(message.updatedAt, options);
                        var keys2;
                        if (message.tasks && (keys2 = Object.keys(message.tasks)).length) {
                            object.tasks = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.tasks[keys2[j]] = $root.fission.workflows.types.TaskStatus.toObject(message.tasks[keys2[j]], options);
                        }
                        if (message.error != null && message.hasOwnProperty("error"))
                            object.error = $root.fission.workflows.types.Error.toObject(message.error, options);
                        return object;
                    };
    
                    /**
                     * Converts this WorkflowStatus to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.WorkflowStatus
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    WorkflowStatus.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    /**
                     * Status enum.
                     * @name fission.workflows.types.WorkflowStatus.Status
                     * @enum {string}
                     * @property {number} PENDING=0 PENDING value
                     * @property {number} READY=2 READY value
                     * @property {number} FAILED=3 FAILED value
                     * @property {number} DELETED=4 DELETED value
                     */
                    WorkflowStatus.Status = (function() {
                        var valuesById = {}, values = Object.create(valuesById);
                        values[valuesById[0] = "PENDING"] = 0;
                        values[valuesById[2] = "READY"] = 2;
                        values[valuesById[3] = "FAILED"] = 3;
                        values[valuesById[4] = "DELETED"] = 4;
                        return values;
                    })();
    
                    return WorkflowStatus;
                })();
    
                types.WorkflowInvocation = (function() {
    
                    /**
                     * Properties of a WorkflowInvocation.
                     * @memberof fission.workflows.types
                     * @interface IWorkflowInvocation
                     * @property {fission.workflows.types.IObjectMetadata|null} [metadata] WorkflowInvocation metadata
                     * @property {fission.workflows.types.IWorkflowInvocationSpec|null} [spec] WorkflowInvocation spec
                     * @property {fission.workflows.types.IWorkflowInvocationStatus|null} [status] WorkflowInvocation status
                     */
    
                    /**
                     * Constructs a new WorkflowInvocation.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a WorkflowInvocation.
                     * @implements IWorkflowInvocation
                     * @constructor
                     * @param {fission.workflows.types.IWorkflowInvocation=} [properties] Properties to set
                     */
                    function WorkflowInvocation(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * WorkflowInvocation metadata.
                     * @member {fission.workflows.types.IObjectMetadata|null|undefined} metadata
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @instance
                     */
                    WorkflowInvocation.prototype.metadata = null;
    
                    /**
                     * WorkflowInvocation spec.
                     * @member {fission.workflows.types.IWorkflowInvocationSpec|null|undefined} spec
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @instance
                     */
                    WorkflowInvocation.prototype.spec = null;
    
                    /**
                     * WorkflowInvocation status.
                     * @member {fission.workflows.types.IWorkflowInvocationStatus|null|undefined} status
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @instance
                     */
                    WorkflowInvocation.prototype.status = null;
    
                    /**
                     * Creates a new WorkflowInvocation instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocation=} [properties] Properties to set
                     * @returns {fission.workflows.types.WorkflowInvocation} WorkflowInvocation instance
                     */
                    WorkflowInvocation.create = function create(properties) {
                        return new WorkflowInvocation(properties);
                    };
    
                    /**
                     * Encodes the specified WorkflowInvocation message. Does not implicitly {@link fission.workflows.types.WorkflowInvocation.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocation} message WorkflowInvocation message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowInvocation.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            $root.fission.workflows.types.ObjectMetadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            $root.fission.workflows.types.WorkflowInvocationSpec.encode(message.spec, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.status != null && message.hasOwnProperty("status"))
                            $root.fission.workflows.types.WorkflowInvocationStatus.encode(message.status, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified WorkflowInvocation message, length delimited. Does not implicitly {@link fission.workflows.types.WorkflowInvocation.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocation} message WorkflowInvocation message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowInvocation.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a WorkflowInvocation message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.WorkflowInvocation} WorkflowInvocation
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowInvocation.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.WorkflowInvocation();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.metadata = $root.fission.workflows.types.ObjectMetadata.decode(reader, reader.uint32());
                                break;
                            case 2:
                                message.spec = $root.fission.workflows.types.WorkflowInvocationSpec.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.status = $root.fission.workflows.types.WorkflowInvocationStatus.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a WorkflowInvocation message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.WorkflowInvocation} WorkflowInvocation
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowInvocation.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a WorkflowInvocation message.
                     * @function verify
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    WorkflowInvocation.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.metadata != null && message.hasOwnProperty("metadata")) {
                            var error = $root.fission.workflows.types.ObjectMetadata.verify(message.metadata);
                            if (error)
                                return "metadata." + error;
                        }
                        if (message.spec != null && message.hasOwnProperty("spec")) {
                            var error = $root.fission.workflows.types.WorkflowInvocationSpec.verify(message.spec);
                            if (error)
                                return "spec." + error;
                        }
                        if (message.status != null && message.hasOwnProperty("status")) {
                            var error = $root.fission.workflows.types.WorkflowInvocationStatus.verify(message.status);
                            if (error)
                                return "status." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a WorkflowInvocation message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.WorkflowInvocation} WorkflowInvocation
                     */
                    WorkflowInvocation.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.WorkflowInvocation)
                            return object;
                        var message = new $root.fission.workflows.types.WorkflowInvocation();
                        if (object.metadata != null) {
                            if (typeof object.metadata !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocation.metadata: object expected");
                            message.metadata = $root.fission.workflows.types.ObjectMetadata.fromObject(object.metadata);
                        }
                        if (object.spec != null) {
                            if (typeof object.spec !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocation.spec: object expected");
                            message.spec = $root.fission.workflows.types.WorkflowInvocationSpec.fromObject(object.spec);
                        }
                        if (object.status != null) {
                            if (typeof object.status !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocation.status: object expected");
                            message.status = $root.fission.workflows.types.WorkflowInvocationStatus.fromObject(object.status);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a WorkflowInvocation message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @static
                     * @param {fission.workflows.types.WorkflowInvocation} message WorkflowInvocation
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    WorkflowInvocation.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.metadata = null;
                            object.spec = null;
                            object.status = null;
                        }
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            object.metadata = $root.fission.workflows.types.ObjectMetadata.toObject(message.metadata, options);
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            object.spec = $root.fission.workflows.types.WorkflowInvocationSpec.toObject(message.spec, options);
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = $root.fission.workflows.types.WorkflowInvocationStatus.toObject(message.status, options);
                        return object;
                    };
    
                    /**
                     * Converts this WorkflowInvocation to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.WorkflowInvocation
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    WorkflowInvocation.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return WorkflowInvocation;
                })();
    
                types.WorkflowInvocationSpec = (function() {
    
                    /**
                     * Properties of a WorkflowInvocationSpec.
                     * @memberof fission.workflows.types
                     * @interface IWorkflowInvocationSpec
                     * @property {string|null} [workflowId] WorkflowInvocationSpec workflowId
                     * @property {Object.<string,fission.workflows.types.ITypedValue>|null} [inputs] WorkflowInvocationSpec inputs
                     * @property {string|null} [parentId] WorkflowInvocationSpec parentId
                     */
    
                    /**
                     * Constructs a new WorkflowInvocationSpec.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a WorkflowInvocationSpec.
                     * @implements IWorkflowInvocationSpec
                     * @constructor
                     * @param {fission.workflows.types.IWorkflowInvocationSpec=} [properties] Properties to set
                     */
                    function WorkflowInvocationSpec(properties) {
                        this.inputs = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * WorkflowInvocationSpec workflowId.
                     * @member {string} workflowId
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @instance
                     */
                    WorkflowInvocationSpec.prototype.workflowId = "";
    
                    /**
                     * WorkflowInvocationSpec inputs.
                     * @member {Object.<string,fission.workflows.types.ITypedValue>} inputs
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @instance
                     */
                    WorkflowInvocationSpec.prototype.inputs = $util.emptyObject;
    
                    /**
                     * WorkflowInvocationSpec parentId.
                     * @member {string} parentId
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @instance
                     */
                    WorkflowInvocationSpec.prototype.parentId = "";
    
                    /**
                     * Creates a new WorkflowInvocationSpec instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocationSpec=} [properties] Properties to set
                     * @returns {fission.workflows.types.WorkflowInvocationSpec} WorkflowInvocationSpec instance
                     */
                    WorkflowInvocationSpec.create = function create(properties) {
                        return new WorkflowInvocationSpec(properties);
                    };
    
                    /**
                     * Encodes the specified WorkflowInvocationSpec message. Does not implicitly {@link fission.workflows.types.WorkflowInvocationSpec.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocationSpec} message WorkflowInvocationSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowInvocationSpec.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.workflowId != null && message.hasOwnProperty("workflowId"))
                            writer.uint32(/* id 1, wireType 2 =*/10).string(message.workflowId);
                        if (message.inputs != null && message.hasOwnProperty("inputs"))
                            for (var keys = Object.keys(message.inputs), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 2, wireType 2 =*/18).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TypedValue.encode(message.inputs[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.parentId != null && message.hasOwnProperty("parentId"))
                            writer.uint32(/* id 3, wireType 2 =*/26).string(message.parentId);
                        return writer;
                    };
    
                    /**
                     * Encodes the specified WorkflowInvocationSpec message, length delimited. Does not implicitly {@link fission.workflows.types.WorkflowInvocationSpec.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocationSpec} message WorkflowInvocationSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowInvocationSpec.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a WorkflowInvocationSpec message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.WorkflowInvocationSpec} WorkflowInvocationSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowInvocationSpec.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.WorkflowInvocationSpec(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.workflowId = reader.string();
                                break;
                            case 2:
                                reader.skip().pos++;
                                if (message.inputs === $util.emptyObject)
                                    message.inputs = {};
                                key = reader.string();
                                reader.pos++;
                                message.inputs[key] = $root.fission.workflows.types.TypedValue.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.parentId = reader.string();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a WorkflowInvocationSpec message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.WorkflowInvocationSpec} WorkflowInvocationSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowInvocationSpec.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a WorkflowInvocationSpec message.
                     * @function verify
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    WorkflowInvocationSpec.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.workflowId != null && message.hasOwnProperty("workflowId"))
                            if (!$util.isString(message.workflowId))
                                return "workflowId: string expected";
                        if (message.inputs != null && message.hasOwnProperty("inputs")) {
                            if (!$util.isObject(message.inputs))
                                return "inputs: object expected";
                            var key = Object.keys(message.inputs);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TypedValue.verify(message.inputs[key[i]]);
                                if (error)
                                    return "inputs." + error;
                            }
                        }
                        if (message.parentId != null && message.hasOwnProperty("parentId"))
                            if (!$util.isString(message.parentId))
                                return "parentId: string expected";
                        return null;
                    };
    
                    /**
                     * Creates a WorkflowInvocationSpec message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.WorkflowInvocationSpec} WorkflowInvocationSpec
                     */
                    WorkflowInvocationSpec.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.WorkflowInvocationSpec)
                            return object;
                        var message = new $root.fission.workflows.types.WorkflowInvocationSpec();
                        if (object.workflowId != null)
                            message.workflowId = String(object.workflowId);
                        if (object.inputs) {
                            if (typeof object.inputs !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocationSpec.inputs: object expected");
                            message.inputs = {};
                            for (var keys = Object.keys(object.inputs), i = 0; i < keys.length; ++i) {
                                if (typeof object.inputs[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.WorkflowInvocationSpec.inputs: object expected");
                                message.inputs[keys[i]] = $root.fission.workflows.types.TypedValue.fromObject(object.inputs[keys[i]]);
                            }
                        }
                        if (object.parentId != null)
                            message.parentId = String(object.parentId);
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a WorkflowInvocationSpec message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @static
                     * @param {fission.workflows.types.WorkflowInvocationSpec} message WorkflowInvocationSpec
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    WorkflowInvocationSpec.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults)
                            object.inputs = {};
                        if (options.defaults) {
                            object.workflowId = "";
                            object.parentId = "";
                        }
                        if (message.workflowId != null && message.hasOwnProperty("workflowId"))
                            object.workflowId = message.workflowId;
                        var keys2;
                        if (message.inputs && (keys2 = Object.keys(message.inputs)).length) {
                            object.inputs = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.inputs[keys2[j]] = $root.fission.workflows.types.TypedValue.toObject(message.inputs[keys2[j]], options);
                        }
                        if (message.parentId != null && message.hasOwnProperty("parentId"))
                            object.parentId = message.parentId;
                        return object;
                    };
    
                    /**
                     * Converts this WorkflowInvocationSpec to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.WorkflowInvocationSpec
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    WorkflowInvocationSpec.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return WorkflowInvocationSpec;
                })();
    
                types.WorkflowInvocationStatus = (function() {
    
                    /**
                     * Properties of a WorkflowInvocationStatus.
                     * @memberof fission.workflows.types
                     * @interface IWorkflowInvocationStatus
                     * @property {fission.workflows.types.WorkflowInvocationStatus.Status|null} [status] WorkflowInvocationStatus status
                     * @property {google.protobuf.ITimestamp|null} [updatedAt] WorkflowInvocationStatus updatedAt
                     * @property {Object.<string,fission.workflows.types.ITaskInvocation>|null} [tasks] WorkflowInvocationStatus tasks
                     * @property {fission.workflows.types.ITypedValue|null} [output] WorkflowInvocationStatus output
                     * @property {Object.<string,fission.workflows.types.ITask>|null} [dynamicTasks] WorkflowInvocationStatus dynamicTasks
                     * @property {fission.workflows.types.IError|null} [error] WorkflowInvocationStatus error
                     */
    
                    /**
                     * Constructs a new WorkflowInvocationStatus.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a WorkflowInvocationStatus.
                     * @implements IWorkflowInvocationStatus
                     * @constructor
                     * @param {fission.workflows.types.IWorkflowInvocationStatus=} [properties] Properties to set
                     */
                    function WorkflowInvocationStatus(properties) {
                        this.tasks = {};
                        this.dynamicTasks = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * WorkflowInvocationStatus status.
                     * @member {fission.workflows.types.WorkflowInvocationStatus.Status} status
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @instance
                     */
                    WorkflowInvocationStatus.prototype.status = 0;
    
                    /**
                     * WorkflowInvocationStatus updatedAt.
                     * @member {google.protobuf.ITimestamp|null|undefined} updatedAt
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @instance
                     */
                    WorkflowInvocationStatus.prototype.updatedAt = null;
    
                    /**
                     * WorkflowInvocationStatus tasks.
                     * @member {Object.<string,fission.workflows.types.ITaskInvocation>} tasks
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @instance
                     */
                    WorkflowInvocationStatus.prototype.tasks = $util.emptyObject;
    
                    /**
                     * WorkflowInvocationStatus output.
                     * @member {fission.workflows.types.ITypedValue|null|undefined} output
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @instance
                     */
                    WorkflowInvocationStatus.prototype.output = null;
    
                    /**
                     * WorkflowInvocationStatus dynamicTasks.
                     * @member {Object.<string,fission.workflows.types.ITask>} dynamicTasks
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @instance
                     */
                    WorkflowInvocationStatus.prototype.dynamicTasks = $util.emptyObject;
    
                    /**
                     * WorkflowInvocationStatus error.
                     * @member {fission.workflows.types.IError|null|undefined} error
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @instance
                     */
                    WorkflowInvocationStatus.prototype.error = null;
    
                    /**
                     * Creates a new WorkflowInvocationStatus instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocationStatus=} [properties] Properties to set
                     * @returns {fission.workflows.types.WorkflowInvocationStatus} WorkflowInvocationStatus instance
                     */
                    WorkflowInvocationStatus.create = function create(properties) {
                        return new WorkflowInvocationStatus(properties);
                    };
    
                    /**
                     * Encodes the specified WorkflowInvocationStatus message. Does not implicitly {@link fission.workflows.types.WorkflowInvocationStatus.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocationStatus} message WorkflowInvocationStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowInvocationStatus.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.status != null && message.hasOwnProperty("status"))
                            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.status);
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            $root.google.protobuf.Timestamp.encode(message.updatedAt, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.tasks != null && message.hasOwnProperty("tasks"))
                            for (var keys = Object.keys(message.tasks), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 3, wireType 2 =*/26).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TaskInvocation.encode(message.tasks[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.output != null && message.hasOwnProperty("output"))
                            $root.fission.workflows.types.TypedValue.encode(message.output, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                        if (message.dynamicTasks != null && message.hasOwnProperty("dynamicTasks"))
                            for (var keys = Object.keys(message.dynamicTasks), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 5, wireType 2 =*/42).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.Task.encode(message.dynamicTasks[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.error != null && message.hasOwnProperty("error"))
                            $root.fission.workflows.types.Error.encode(message.error, writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified WorkflowInvocationStatus message, length delimited. Does not implicitly {@link fission.workflows.types.WorkflowInvocationStatus.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {fission.workflows.types.IWorkflowInvocationStatus} message WorkflowInvocationStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    WorkflowInvocationStatus.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a WorkflowInvocationStatus message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.WorkflowInvocationStatus} WorkflowInvocationStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowInvocationStatus.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.WorkflowInvocationStatus(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.status = reader.int32();
                                break;
                            case 2:
                                message.updatedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                                break;
                            case 3:
                                reader.skip().pos++;
                                if (message.tasks === $util.emptyObject)
                                    message.tasks = {};
                                key = reader.string();
                                reader.pos++;
                                message.tasks[key] = $root.fission.workflows.types.TaskInvocation.decode(reader, reader.uint32());
                                break;
                            case 4:
                                message.output = $root.fission.workflows.types.TypedValue.decode(reader, reader.uint32());
                                break;
                            case 5:
                                reader.skip().pos++;
                                if (message.dynamicTasks === $util.emptyObject)
                                    message.dynamicTasks = {};
                                key = reader.string();
                                reader.pos++;
                                message.dynamicTasks[key] = $root.fission.workflows.types.Task.decode(reader, reader.uint32());
                                break;
                            case 6:
                                message.error = $root.fission.workflows.types.Error.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a WorkflowInvocationStatus message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.WorkflowInvocationStatus} WorkflowInvocationStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    WorkflowInvocationStatus.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a WorkflowInvocationStatus message.
                     * @function verify
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    WorkflowInvocationStatus.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.status != null && message.hasOwnProperty("status"))
                            switch (message.status) {
                            default:
                                return "status: enum value expected";
                            case 0:
                            case 1:
                            case 2:
                            case 3:
                            case 4:
                            case 5:
                                break;
                            }
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt")) {
                            var error = $root.google.protobuf.Timestamp.verify(message.updatedAt);
                            if (error)
                                return "updatedAt." + error;
                        }
                        if (message.tasks != null && message.hasOwnProperty("tasks")) {
                            if (!$util.isObject(message.tasks))
                                return "tasks: object expected";
                            var key = Object.keys(message.tasks);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TaskInvocation.verify(message.tasks[key[i]]);
                                if (error)
                                    return "tasks." + error;
                            }
                        }
                        if (message.output != null && message.hasOwnProperty("output")) {
                            var error = $root.fission.workflows.types.TypedValue.verify(message.output);
                            if (error)
                                return "output." + error;
                        }
                        if (message.dynamicTasks != null && message.hasOwnProperty("dynamicTasks")) {
                            if (!$util.isObject(message.dynamicTasks))
                                return "dynamicTasks: object expected";
                            var key = Object.keys(message.dynamicTasks);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.Task.verify(message.dynamicTasks[key[i]]);
                                if (error)
                                    return "dynamicTasks." + error;
                            }
                        }
                        if (message.error != null && message.hasOwnProperty("error")) {
                            var error = $root.fission.workflows.types.Error.verify(message.error);
                            if (error)
                                return "error." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a WorkflowInvocationStatus message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.WorkflowInvocationStatus} WorkflowInvocationStatus
                     */
                    WorkflowInvocationStatus.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.WorkflowInvocationStatus)
                            return object;
                        var message = new $root.fission.workflows.types.WorkflowInvocationStatus();
                        switch (object.status) {
                        case "UNKNOWN":
                        case 0:
                            message.status = 0;
                            break;
                        case "SCHEDULED":
                        case 1:
                            message.status = 1;
                            break;
                        case "IN_PROGRESS":
                        case 2:
                            message.status = 2;
                            break;
                        case "SUCCEEDED":
                        case 3:
                            message.status = 3;
                            break;
                        case "FAILED":
                        case 4:
                            message.status = 4;
                            break;
                        case "ABORTED":
                        case 5:
                            message.status = 5;
                            break;
                        }
                        if (object.updatedAt != null) {
                            if (typeof object.updatedAt !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocationStatus.updatedAt: object expected");
                            message.updatedAt = $root.google.protobuf.Timestamp.fromObject(object.updatedAt);
                        }
                        if (object.tasks) {
                            if (typeof object.tasks !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocationStatus.tasks: object expected");
                            message.tasks = {};
                            for (var keys = Object.keys(object.tasks), i = 0; i < keys.length; ++i) {
                                if (typeof object.tasks[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.WorkflowInvocationStatus.tasks: object expected");
                                message.tasks[keys[i]] = $root.fission.workflows.types.TaskInvocation.fromObject(object.tasks[keys[i]]);
                            }
                        }
                        if (object.output != null) {
                            if (typeof object.output !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocationStatus.output: object expected");
                            message.output = $root.fission.workflows.types.TypedValue.fromObject(object.output);
                        }
                        if (object.dynamicTasks) {
                            if (typeof object.dynamicTasks !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocationStatus.dynamicTasks: object expected");
                            message.dynamicTasks = {};
                            for (var keys = Object.keys(object.dynamicTasks), i = 0; i < keys.length; ++i) {
                                if (typeof object.dynamicTasks[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.WorkflowInvocationStatus.dynamicTasks: object expected");
                                message.dynamicTasks[keys[i]] = $root.fission.workflows.types.Task.fromObject(object.dynamicTasks[keys[i]]);
                            }
                        }
                        if (object.error != null) {
                            if (typeof object.error !== "object")
                                throw TypeError(".fission.workflows.types.WorkflowInvocationStatus.error: object expected");
                            message.error = $root.fission.workflows.types.Error.fromObject(object.error);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a WorkflowInvocationStatus message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @static
                     * @param {fission.workflows.types.WorkflowInvocationStatus} message WorkflowInvocationStatus
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    WorkflowInvocationStatus.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults) {
                            object.tasks = {};
                            object.dynamicTasks = {};
                        }
                        if (options.defaults) {
                            object.status = options.enums === String ? "UNKNOWN" : 0;
                            object.updatedAt = null;
                            object.output = null;
                            object.error = null;
                        }
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = options.enums === String ? $root.fission.workflows.types.WorkflowInvocationStatus.Status[message.status] : message.status;
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            object.updatedAt = $root.google.protobuf.Timestamp.toObject(message.updatedAt, options);
                        var keys2;
                        if (message.tasks && (keys2 = Object.keys(message.tasks)).length) {
                            object.tasks = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.tasks[keys2[j]] = $root.fission.workflows.types.TaskInvocation.toObject(message.tasks[keys2[j]], options);
                        }
                        if (message.output != null && message.hasOwnProperty("output"))
                            object.output = $root.fission.workflows.types.TypedValue.toObject(message.output, options);
                        if (message.dynamicTasks && (keys2 = Object.keys(message.dynamicTasks)).length) {
                            object.dynamicTasks = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.dynamicTasks[keys2[j]] = $root.fission.workflows.types.Task.toObject(message.dynamicTasks[keys2[j]], options);
                        }
                        if (message.error != null && message.hasOwnProperty("error"))
                            object.error = $root.fission.workflows.types.Error.toObject(message.error, options);
                        return object;
                    };
    
                    /**
                     * Converts this WorkflowInvocationStatus to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.WorkflowInvocationStatus
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    WorkflowInvocationStatus.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    /**
                     * Status enum.
                     * @name fission.workflows.types.WorkflowInvocationStatus.Status
                     * @enum {string}
                     * @property {number} UNKNOWN=0 UNKNOWN value
                     * @property {number} SCHEDULED=1 SCHEDULED value
                     * @property {number} IN_PROGRESS=2 IN_PROGRESS value
                     * @property {number} SUCCEEDED=3 SUCCEEDED value
                     * @property {number} FAILED=4 FAILED value
                     * @property {number} ABORTED=5 ABORTED value
                     */
                    WorkflowInvocationStatus.Status = (function() {
                        var valuesById = {}, values = Object.create(valuesById);
                        values[valuesById[0] = "UNKNOWN"] = 0;
                        values[valuesById[1] = "SCHEDULED"] = 1;
                        values[valuesById[2] = "IN_PROGRESS"] = 2;
                        values[valuesById[3] = "SUCCEEDED"] = 3;
                        values[valuesById[4] = "FAILED"] = 4;
                        values[valuesById[5] = "ABORTED"] = 5;
                        return values;
                    })();
    
                    return WorkflowInvocationStatus;
                })();
    
                types.DependencyConfig = (function() {
    
                    /**
                     * Properties of a DependencyConfig.
                     * @memberof fission.workflows.types
                     * @interface IDependencyConfig
                     * @property {Object.<string,fission.workflows.types.ITaskDependencyParameters>|null} [requires] DependencyConfig requires
                     * @property {number|null} [await] DependencyConfig await
                     */
    
                    /**
                     * Constructs a new DependencyConfig.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a DependencyConfig.
                     * @implements IDependencyConfig
                     * @constructor
                     * @param {fission.workflows.types.IDependencyConfig=} [properties] Properties to set
                     */
                    function DependencyConfig(properties) {
                        this.requires = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * DependencyConfig requires.
                     * @member {Object.<string,fission.workflows.types.ITaskDependencyParameters>} requires
                     * @memberof fission.workflows.types.DependencyConfig
                     * @instance
                     */
                    DependencyConfig.prototype.requires = $util.emptyObject;
    
                    /**
                     * DependencyConfig await.
                     * @member {number} await
                     * @memberof fission.workflows.types.DependencyConfig
                     * @instance
                     */
                    DependencyConfig.prototype.await = 0;
    
                    /**
                     * Creates a new DependencyConfig instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {fission.workflows.types.IDependencyConfig=} [properties] Properties to set
                     * @returns {fission.workflows.types.DependencyConfig} DependencyConfig instance
                     */
                    DependencyConfig.create = function create(properties) {
                        return new DependencyConfig(properties);
                    };
    
                    /**
                     * Encodes the specified DependencyConfig message. Does not implicitly {@link fission.workflows.types.DependencyConfig.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {fission.workflows.types.IDependencyConfig} message DependencyConfig message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    DependencyConfig.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.requires != null && message.hasOwnProperty("requires"))
                            for (var keys = Object.keys(message.requires), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 1, wireType 2 =*/10).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TaskDependencyParameters.encode(message.requires[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.await != null && message.hasOwnProperty("await"))
                            writer.uint32(/* id 2, wireType 0 =*/16).int32(message.await);
                        return writer;
                    };
    
                    /**
                     * Encodes the specified DependencyConfig message, length delimited. Does not implicitly {@link fission.workflows.types.DependencyConfig.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {fission.workflows.types.IDependencyConfig} message DependencyConfig message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    DependencyConfig.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a DependencyConfig message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.DependencyConfig} DependencyConfig
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    DependencyConfig.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.DependencyConfig(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                reader.skip().pos++;
                                if (message.requires === $util.emptyObject)
                                    message.requires = {};
                                key = reader.string();
                                reader.pos++;
                                message.requires[key] = $root.fission.workflows.types.TaskDependencyParameters.decode(reader, reader.uint32());
                                break;
                            case 2:
                                message.await = reader.int32();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a DependencyConfig message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.DependencyConfig} DependencyConfig
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    DependencyConfig.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a DependencyConfig message.
                     * @function verify
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    DependencyConfig.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.requires != null && message.hasOwnProperty("requires")) {
                            if (!$util.isObject(message.requires))
                                return "requires: object expected";
                            var key = Object.keys(message.requires);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TaskDependencyParameters.verify(message.requires[key[i]]);
                                if (error)
                                    return "requires." + error;
                            }
                        }
                        if (message.await != null && message.hasOwnProperty("await"))
                            if (!$util.isInteger(message.await))
                                return "await: integer expected";
                        return null;
                    };
    
                    /**
                     * Creates a DependencyConfig message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.DependencyConfig} DependencyConfig
                     */
                    DependencyConfig.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.DependencyConfig)
                            return object;
                        var message = new $root.fission.workflows.types.DependencyConfig();
                        if (object.requires) {
                            if (typeof object.requires !== "object")
                                throw TypeError(".fission.workflows.types.DependencyConfig.requires: object expected");
                            message.requires = {};
                            for (var keys = Object.keys(object.requires), i = 0; i < keys.length; ++i) {
                                if (typeof object.requires[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.DependencyConfig.requires: object expected");
                                message.requires[keys[i]] = $root.fission.workflows.types.TaskDependencyParameters.fromObject(object.requires[keys[i]]);
                            }
                        }
                        if (object.await != null)
                            message.await = object.await | 0;
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a DependencyConfig message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.DependencyConfig
                     * @static
                     * @param {fission.workflows.types.DependencyConfig} message DependencyConfig
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    DependencyConfig.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults)
                            object.requires = {};
                        if (options.defaults)
                            object.await = 0;
                        var keys2;
                        if (message.requires && (keys2 = Object.keys(message.requires)).length) {
                            object.requires = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.requires[keys2[j]] = $root.fission.workflows.types.TaskDependencyParameters.toObject(message.requires[keys2[j]], options);
                        }
                        if (message.await != null && message.hasOwnProperty("await"))
                            object.await = message.await;
                        return object;
                    };
    
                    /**
                     * Converts this DependencyConfig to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.DependencyConfig
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    DependencyConfig.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return DependencyConfig;
                })();
    
                types.Task = (function() {
    
                    /**
                     * Properties of a Task.
                     * @memberof fission.workflows.types
                     * @interface ITask
                     * @property {fission.workflows.types.IObjectMetadata|null} [metadata] Task metadata
                     * @property {fission.workflows.types.ITaskSpec|null} [spec] Task spec
                     * @property {fission.workflows.types.ITaskStatus|null} [status] Task status
                     */
    
                    /**
                     * Constructs a new Task.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a Task.
                     * @implements ITask
                     * @constructor
                     * @param {fission.workflows.types.ITask=} [properties] Properties to set
                     */
                    function Task(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * Task metadata.
                     * @member {fission.workflows.types.IObjectMetadata|null|undefined} metadata
                     * @memberof fission.workflows.types.Task
                     * @instance
                     */
                    Task.prototype.metadata = null;
    
                    /**
                     * Task spec.
                     * @member {fission.workflows.types.ITaskSpec|null|undefined} spec
                     * @memberof fission.workflows.types.Task
                     * @instance
                     */
                    Task.prototype.spec = null;
    
                    /**
                     * Task status.
                     * @member {fission.workflows.types.ITaskStatus|null|undefined} status
                     * @memberof fission.workflows.types.Task
                     * @instance
                     */
                    Task.prototype.status = null;
    
                    /**
                     * Creates a new Task instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {fission.workflows.types.ITask=} [properties] Properties to set
                     * @returns {fission.workflows.types.Task} Task instance
                     */
                    Task.create = function create(properties) {
                        return new Task(properties);
                    };
    
                    /**
                     * Encodes the specified Task message. Does not implicitly {@link fission.workflows.types.Task.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {fission.workflows.types.ITask} message Task message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    Task.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            $root.fission.workflows.types.ObjectMetadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            $root.fission.workflows.types.TaskSpec.encode(message.spec, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.status != null && message.hasOwnProperty("status"))
                            $root.fission.workflows.types.TaskStatus.encode(message.status, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified Task message, length delimited. Does not implicitly {@link fission.workflows.types.Task.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {fission.workflows.types.ITask} message Task message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    Task.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a Task message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.Task} Task
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    Task.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.Task();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.metadata = $root.fission.workflows.types.ObjectMetadata.decode(reader, reader.uint32());
                                break;
                            case 2:
                                message.spec = $root.fission.workflows.types.TaskSpec.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.status = $root.fission.workflows.types.TaskStatus.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a Task message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.Task} Task
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    Task.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a Task message.
                     * @function verify
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    Task.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.metadata != null && message.hasOwnProperty("metadata")) {
                            var error = $root.fission.workflows.types.ObjectMetadata.verify(message.metadata);
                            if (error)
                                return "metadata." + error;
                        }
                        if (message.spec != null && message.hasOwnProperty("spec")) {
                            var error = $root.fission.workflows.types.TaskSpec.verify(message.spec);
                            if (error)
                                return "spec." + error;
                        }
                        if (message.status != null && message.hasOwnProperty("status")) {
                            var error = $root.fission.workflows.types.TaskStatus.verify(message.status);
                            if (error)
                                return "status." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a Task message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.Task} Task
                     */
                    Task.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.Task)
                            return object;
                        var message = new $root.fission.workflows.types.Task();
                        if (object.metadata != null) {
                            if (typeof object.metadata !== "object")
                                throw TypeError(".fission.workflows.types.Task.metadata: object expected");
                            message.metadata = $root.fission.workflows.types.ObjectMetadata.fromObject(object.metadata);
                        }
                        if (object.spec != null) {
                            if (typeof object.spec !== "object")
                                throw TypeError(".fission.workflows.types.Task.spec: object expected");
                            message.spec = $root.fission.workflows.types.TaskSpec.fromObject(object.spec);
                        }
                        if (object.status != null) {
                            if (typeof object.status !== "object")
                                throw TypeError(".fission.workflows.types.Task.status: object expected");
                            message.status = $root.fission.workflows.types.TaskStatus.fromObject(object.status);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a Task message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.Task
                     * @static
                     * @param {fission.workflows.types.Task} message Task
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    Task.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.metadata = null;
                            object.spec = null;
                            object.status = null;
                        }
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            object.metadata = $root.fission.workflows.types.ObjectMetadata.toObject(message.metadata, options);
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            object.spec = $root.fission.workflows.types.TaskSpec.toObject(message.spec, options);
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = $root.fission.workflows.types.TaskStatus.toObject(message.status, options);
                        return object;
                    };
    
                    /**
                     * Converts this Task to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.Task
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    Task.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return Task;
                })();
    
                types.TaskSpec = (function() {
    
                    /**
                     * Properties of a TaskSpec.
                     * @memberof fission.workflows.types
                     * @interface ITaskSpec
                     * @property {string|null} [functionRef] TaskSpec functionRef
                     * @property {Object.<string,fission.workflows.types.ITypedValue>|null} [inputs] TaskSpec inputs
                     * @property {Object.<string,fission.workflows.types.ITaskDependencyParameters>|null} [requires] TaskSpec requires
                     * @property {number|null} [await] TaskSpec await
                     * @property {fission.workflows.types.ITypedValue|null} [output] TaskSpec output
                     */
    
                    /**
                     * Constructs a new TaskSpec.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TaskSpec.
                     * @implements ITaskSpec
                     * @constructor
                     * @param {fission.workflows.types.ITaskSpec=} [properties] Properties to set
                     */
                    function TaskSpec(properties) {
                        this.inputs = {};
                        this.requires = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TaskSpec functionRef.
                     * @member {string} functionRef
                     * @memberof fission.workflows.types.TaskSpec
                     * @instance
                     */
                    TaskSpec.prototype.functionRef = "";
    
                    /**
                     * TaskSpec inputs.
                     * @member {Object.<string,fission.workflows.types.ITypedValue>} inputs
                     * @memberof fission.workflows.types.TaskSpec
                     * @instance
                     */
                    TaskSpec.prototype.inputs = $util.emptyObject;
    
                    /**
                     * TaskSpec requires.
                     * @member {Object.<string,fission.workflows.types.ITaskDependencyParameters>} requires
                     * @memberof fission.workflows.types.TaskSpec
                     * @instance
                     */
                    TaskSpec.prototype.requires = $util.emptyObject;
    
                    /**
                     * TaskSpec await.
                     * @member {number} await
                     * @memberof fission.workflows.types.TaskSpec
                     * @instance
                     */
                    TaskSpec.prototype.await = 0;
    
                    /**
                     * TaskSpec output.
                     * @member {fission.workflows.types.ITypedValue|null|undefined} output
                     * @memberof fission.workflows.types.TaskSpec
                     * @instance
                     */
                    TaskSpec.prototype.output = null;
    
                    /**
                     * Creates a new TaskSpec instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {fission.workflows.types.ITaskSpec=} [properties] Properties to set
                     * @returns {fission.workflows.types.TaskSpec} TaskSpec instance
                     */
                    TaskSpec.create = function create(properties) {
                        return new TaskSpec(properties);
                    };
    
                    /**
                     * Encodes the specified TaskSpec message. Does not implicitly {@link fission.workflows.types.TaskSpec.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {fission.workflows.types.ITaskSpec} message TaskSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskSpec.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.functionRef != null && message.hasOwnProperty("functionRef"))
                            writer.uint32(/* id 1, wireType 2 =*/10).string(message.functionRef);
                        if (message.inputs != null && message.hasOwnProperty("inputs"))
                            for (var keys = Object.keys(message.inputs), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 2, wireType 2 =*/18).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TypedValue.encode(message.inputs[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.requires != null && message.hasOwnProperty("requires"))
                            for (var keys = Object.keys(message.requires), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 3, wireType 2 =*/26).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TaskDependencyParameters.encode(message.requires[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.await != null && message.hasOwnProperty("await"))
                            writer.uint32(/* id 4, wireType 0 =*/32).int32(message.await);
                        if (message.output != null && message.hasOwnProperty("output"))
                            $root.fission.workflows.types.TypedValue.encode(message.output, writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TaskSpec message, length delimited. Does not implicitly {@link fission.workflows.types.TaskSpec.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {fission.workflows.types.ITaskSpec} message TaskSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskSpec.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TaskSpec message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TaskSpec} TaskSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskSpec.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TaskSpec(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.functionRef = reader.string();
                                break;
                            case 2:
                                reader.skip().pos++;
                                if (message.inputs === $util.emptyObject)
                                    message.inputs = {};
                                key = reader.string();
                                reader.pos++;
                                message.inputs[key] = $root.fission.workflows.types.TypedValue.decode(reader, reader.uint32());
                                break;
                            case 3:
                                reader.skip().pos++;
                                if (message.requires === $util.emptyObject)
                                    message.requires = {};
                                key = reader.string();
                                reader.pos++;
                                message.requires[key] = $root.fission.workflows.types.TaskDependencyParameters.decode(reader, reader.uint32());
                                break;
                            case 4:
                                message.await = reader.int32();
                                break;
                            case 5:
                                message.output = $root.fission.workflows.types.TypedValue.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TaskSpec message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TaskSpec} TaskSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskSpec.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TaskSpec message.
                     * @function verify
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TaskSpec.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.functionRef != null && message.hasOwnProperty("functionRef"))
                            if (!$util.isString(message.functionRef))
                                return "functionRef: string expected";
                        if (message.inputs != null && message.hasOwnProperty("inputs")) {
                            if (!$util.isObject(message.inputs))
                                return "inputs: object expected";
                            var key = Object.keys(message.inputs);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TypedValue.verify(message.inputs[key[i]]);
                                if (error)
                                    return "inputs." + error;
                            }
                        }
                        if (message.requires != null && message.hasOwnProperty("requires")) {
                            if (!$util.isObject(message.requires))
                                return "requires: object expected";
                            var key = Object.keys(message.requires);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TaskDependencyParameters.verify(message.requires[key[i]]);
                                if (error)
                                    return "requires." + error;
                            }
                        }
                        if (message.await != null && message.hasOwnProperty("await"))
                            if (!$util.isInteger(message.await))
                                return "await: integer expected";
                        if (message.output != null && message.hasOwnProperty("output")) {
                            var error = $root.fission.workflows.types.TypedValue.verify(message.output);
                            if (error)
                                return "output." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a TaskSpec message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TaskSpec} TaskSpec
                     */
                    TaskSpec.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TaskSpec)
                            return object;
                        var message = new $root.fission.workflows.types.TaskSpec();
                        if (object.functionRef != null)
                            message.functionRef = String(object.functionRef);
                        if (object.inputs) {
                            if (typeof object.inputs !== "object")
                                throw TypeError(".fission.workflows.types.TaskSpec.inputs: object expected");
                            message.inputs = {};
                            for (var keys = Object.keys(object.inputs), i = 0; i < keys.length; ++i) {
                                if (typeof object.inputs[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.TaskSpec.inputs: object expected");
                                message.inputs[keys[i]] = $root.fission.workflows.types.TypedValue.fromObject(object.inputs[keys[i]]);
                            }
                        }
                        if (object.requires) {
                            if (typeof object.requires !== "object")
                                throw TypeError(".fission.workflows.types.TaskSpec.requires: object expected");
                            message.requires = {};
                            for (var keys = Object.keys(object.requires), i = 0; i < keys.length; ++i) {
                                if (typeof object.requires[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.TaskSpec.requires: object expected");
                                message.requires[keys[i]] = $root.fission.workflows.types.TaskDependencyParameters.fromObject(object.requires[keys[i]]);
                            }
                        }
                        if (object.await != null)
                            message.await = object.await | 0;
                        if (object.output != null) {
                            if (typeof object.output !== "object")
                                throw TypeError(".fission.workflows.types.TaskSpec.output: object expected");
                            message.output = $root.fission.workflows.types.TypedValue.fromObject(object.output);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TaskSpec message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TaskSpec
                     * @static
                     * @param {fission.workflows.types.TaskSpec} message TaskSpec
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TaskSpec.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults) {
                            object.inputs = {};
                            object.requires = {};
                        }
                        if (options.defaults) {
                            object.functionRef = "";
                            object.await = 0;
                            object.output = null;
                        }
                        if (message.functionRef != null && message.hasOwnProperty("functionRef"))
                            object.functionRef = message.functionRef;
                        var keys2;
                        if (message.inputs && (keys2 = Object.keys(message.inputs)).length) {
                            object.inputs = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.inputs[keys2[j]] = $root.fission.workflows.types.TypedValue.toObject(message.inputs[keys2[j]], options);
                        }
                        if (message.requires && (keys2 = Object.keys(message.requires)).length) {
                            object.requires = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.requires[keys2[j]] = $root.fission.workflows.types.TaskDependencyParameters.toObject(message.requires[keys2[j]], options);
                        }
                        if (message.await != null && message.hasOwnProperty("await"))
                            object.await = message.await;
                        if (message.output != null && message.hasOwnProperty("output"))
                            object.output = $root.fission.workflows.types.TypedValue.toObject(message.output, options);
                        return object;
                    };
    
                    /**
                     * Converts this TaskSpec to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TaskSpec
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TaskSpec.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return TaskSpec;
                })();
    
                types.TaskStatus = (function() {
    
                    /**
                     * Properties of a TaskStatus.
                     * @memberof fission.workflows.types
                     * @interface ITaskStatus
                     * @property {fission.workflows.types.TaskStatus.Status|null} [status] TaskStatus status
                     * @property {google.protobuf.ITimestamp|null} [updatedAt] TaskStatus updatedAt
                     * @property {fission.workflows.types.IFnRef|null} [fnRef] TaskStatus fnRef
                     * @property {fission.workflows.types.IError|null} [error] TaskStatus error
                     */
    
                    /**
                     * Constructs a new TaskStatus.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TaskStatus.
                     * @implements ITaskStatus
                     * @constructor
                     * @param {fission.workflows.types.ITaskStatus=} [properties] Properties to set
                     */
                    function TaskStatus(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TaskStatus status.
                     * @member {fission.workflows.types.TaskStatus.Status} status
                     * @memberof fission.workflows.types.TaskStatus
                     * @instance
                     */
                    TaskStatus.prototype.status = 0;
    
                    /**
                     * TaskStatus updatedAt.
                     * @member {google.protobuf.ITimestamp|null|undefined} updatedAt
                     * @memberof fission.workflows.types.TaskStatus
                     * @instance
                     */
                    TaskStatus.prototype.updatedAt = null;
    
                    /**
                     * TaskStatus fnRef.
                     * @member {fission.workflows.types.IFnRef|null|undefined} fnRef
                     * @memberof fission.workflows.types.TaskStatus
                     * @instance
                     */
                    TaskStatus.prototype.fnRef = null;
    
                    /**
                     * TaskStatus error.
                     * @member {fission.workflows.types.IError|null|undefined} error
                     * @memberof fission.workflows.types.TaskStatus
                     * @instance
                     */
                    TaskStatus.prototype.error = null;
    
                    /**
                     * Creates a new TaskStatus instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {fission.workflows.types.ITaskStatus=} [properties] Properties to set
                     * @returns {fission.workflows.types.TaskStatus} TaskStatus instance
                     */
                    TaskStatus.create = function create(properties) {
                        return new TaskStatus(properties);
                    };
    
                    /**
                     * Encodes the specified TaskStatus message. Does not implicitly {@link fission.workflows.types.TaskStatus.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {fission.workflows.types.ITaskStatus} message TaskStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskStatus.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.status != null && message.hasOwnProperty("status"))
                            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.status);
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            $root.google.protobuf.Timestamp.encode(message.updatedAt, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.fnRef != null && message.hasOwnProperty("fnRef"))
                            $root.fission.workflows.types.FnRef.encode(message.fnRef, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                        if (message.error != null && message.hasOwnProperty("error"))
                            $root.fission.workflows.types.Error.encode(message.error, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TaskStatus message, length delimited. Does not implicitly {@link fission.workflows.types.TaskStatus.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {fission.workflows.types.ITaskStatus} message TaskStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskStatus.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TaskStatus message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TaskStatus} TaskStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskStatus.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TaskStatus();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.status = reader.int32();
                                break;
                            case 2:
                                message.updatedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.fnRef = $root.fission.workflows.types.FnRef.decode(reader, reader.uint32());
                                break;
                            case 4:
                                message.error = $root.fission.workflows.types.Error.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TaskStatus message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TaskStatus} TaskStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskStatus.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TaskStatus message.
                     * @function verify
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TaskStatus.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.status != null && message.hasOwnProperty("status"))
                            switch (message.status) {
                            default:
                                return "status: enum value expected";
                            case 0:
                            case 1:
                            case 2:
                                break;
                            }
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt")) {
                            var error = $root.google.protobuf.Timestamp.verify(message.updatedAt);
                            if (error)
                                return "updatedAt." + error;
                        }
                        if (message.fnRef != null && message.hasOwnProperty("fnRef")) {
                            var error = $root.fission.workflows.types.FnRef.verify(message.fnRef);
                            if (error)
                                return "fnRef." + error;
                        }
                        if (message.error != null && message.hasOwnProperty("error")) {
                            var error = $root.fission.workflows.types.Error.verify(message.error);
                            if (error)
                                return "error." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a TaskStatus message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TaskStatus} TaskStatus
                     */
                    TaskStatus.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TaskStatus)
                            return object;
                        var message = new $root.fission.workflows.types.TaskStatus();
                        switch (object.status) {
                        case "STARTED":
                        case 0:
                            message.status = 0;
                            break;
                        case "READY":
                        case 1:
                            message.status = 1;
                            break;
                        case "FAILED":
                        case 2:
                            message.status = 2;
                            break;
                        }
                        if (object.updatedAt != null) {
                            if (typeof object.updatedAt !== "object")
                                throw TypeError(".fission.workflows.types.TaskStatus.updatedAt: object expected");
                            message.updatedAt = $root.google.protobuf.Timestamp.fromObject(object.updatedAt);
                        }
                        if (object.fnRef != null) {
                            if (typeof object.fnRef !== "object")
                                throw TypeError(".fission.workflows.types.TaskStatus.fnRef: object expected");
                            message.fnRef = $root.fission.workflows.types.FnRef.fromObject(object.fnRef);
                        }
                        if (object.error != null) {
                            if (typeof object.error !== "object")
                                throw TypeError(".fission.workflows.types.TaskStatus.error: object expected");
                            message.error = $root.fission.workflows.types.Error.fromObject(object.error);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TaskStatus message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TaskStatus
                     * @static
                     * @param {fission.workflows.types.TaskStatus} message TaskStatus
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TaskStatus.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.status = options.enums === String ? "STARTED" : 0;
                            object.updatedAt = null;
                            object.fnRef = null;
                            object.error = null;
                        }
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = options.enums === String ? $root.fission.workflows.types.TaskStatus.Status[message.status] : message.status;
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            object.updatedAt = $root.google.protobuf.Timestamp.toObject(message.updatedAt, options);
                        if (message.fnRef != null && message.hasOwnProperty("fnRef"))
                            object.fnRef = $root.fission.workflows.types.FnRef.toObject(message.fnRef, options);
                        if (message.error != null && message.hasOwnProperty("error"))
                            object.error = $root.fission.workflows.types.Error.toObject(message.error, options);
                        return object;
                    };
    
                    /**
                     * Converts this TaskStatus to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TaskStatus
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TaskStatus.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    /**
                     * Status enum.
                     * @name fission.workflows.types.TaskStatus.Status
                     * @enum {string}
                     * @property {number} STARTED=0 STARTED value
                     * @property {number} READY=1 READY value
                     * @property {number} FAILED=2 FAILED value
                     */
                    TaskStatus.Status = (function() {
                        var valuesById = {}, values = Object.create(valuesById);
                        values[valuesById[0] = "STARTED"] = 0;
                        values[valuesById[1] = "READY"] = 1;
                        values[valuesById[2] = "FAILED"] = 2;
                        return values;
                    })();
    
                    return TaskStatus;
                })();
    
                types.TaskDependencyParameters = (function() {
    
                    /**
                     * Properties of a TaskDependencyParameters.
                     * @memberof fission.workflows.types
                     * @interface ITaskDependencyParameters
                     * @property {fission.workflows.types.TaskDependencyParameters.DependencyType|null} [type] TaskDependencyParameters type
                     * @property {string|null} [alias] TaskDependencyParameters alias
                     */
    
                    /**
                     * Constructs a new TaskDependencyParameters.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TaskDependencyParameters.
                     * @implements ITaskDependencyParameters
                     * @constructor
                     * @param {fission.workflows.types.ITaskDependencyParameters=} [properties] Properties to set
                     */
                    function TaskDependencyParameters(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TaskDependencyParameters type.
                     * @member {fission.workflows.types.TaskDependencyParameters.DependencyType} type
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @instance
                     */
                    TaskDependencyParameters.prototype.type = 0;
    
                    /**
                     * TaskDependencyParameters alias.
                     * @member {string} alias
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @instance
                     */
                    TaskDependencyParameters.prototype.alias = "";
    
                    /**
                     * Creates a new TaskDependencyParameters instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {fission.workflows.types.ITaskDependencyParameters=} [properties] Properties to set
                     * @returns {fission.workflows.types.TaskDependencyParameters} TaskDependencyParameters instance
                     */
                    TaskDependencyParameters.create = function create(properties) {
                        return new TaskDependencyParameters(properties);
                    };
    
                    /**
                     * Encodes the specified TaskDependencyParameters message. Does not implicitly {@link fission.workflows.types.TaskDependencyParameters.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {fission.workflows.types.ITaskDependencyParameters} message TaskDependencyParameters message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskDependencyParameters.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.type != null && message.hasOwnProperty("type"))
                            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.type);
                        if (message.alias != null && message.hasOwnProperty("alias"))
                            writer.uint32(/* id 2, wireType 2 =*/18).string(message.alias);
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TaskDependencyParameters message, length delimited. Does not implicitly {@link fission.workflows.types.TaskDependencyParameters.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {fission.workflows.types.ITaskDependencyParameters} message TaskDependencyParameters message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskDependencyParameters.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TaskDependencyParameters message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TaskDependencyParameters} TaskDependencyParameters
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskDependencyParameters.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TaskDependencyParameters();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.type = reader.int32();
                                break;
                            case 2:
                                message.alias = reader.string();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TaskDependencyParameters message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TaskDependencyParameters} TaskDependencyParameters
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskDependencyParameters.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TaskDependencyParameters message.
                     * @function verify
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TaskDependencyParameters.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.type != null && message.hasOwnProperty("type"))
                            switch (message.type) {
                            default:
                                return "type: enum value expected";
                            case 0:
                            case 1:
                            case 2:
                                break;
                            }
                        if (message.alias != null && message.hasOwnProperty("alias"))
                            if (!$util.isString(message.alias))
                                return "alias: string expected";
                        return null;
                    };
    
                    /**
                     * Creates a TaskDependencyParameters message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TaskDependencyParameters} TaskDependencyParameters
                     */
                    TaskDependencyParameters.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TaskDependencyParameters)
                            return object;
                        var message = new $root.fission.workflows.types.TaskDependencyParameters();
                        switch (object.type) {
                        case "DATA":
                        case 0:
                            message.type = 0;
                            break;
                        case "CONTROL":
                        case 1:
                            message.type = 1;
                            break;
                        case "DYNAMIC_OUTPUT":
                        case 2:
                            message.type = 2;
                            break;
                        }
                        if (object.alias != null)
                            message.alias = String(object.alias);
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TaskDependencyParameters message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @static
                     * @param {fission.workflows.types.TaskDependencyParameters} message TaskDependencyParameters
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TaskDependencyParameters.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.type = options.enums === String ? "DATA" : 0;
                            object.alias = "";
                        }
                        if (message.type != null && message.hasOwnProperty("type"))
                            object.type = options.enums === String ? $root.fission.workflows.types.TaskDependencyParameters.DependencyType[message.type] : message.type;
                        if (message.alias != null && message.hasOwnProperty("alias"))
                            object.alias = message.alias;
                        return object;
                    };
    
                    /**
                     * Converts this TaskDependencyParameters to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TaskDependencyParameters
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TaskDependencyParameters.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    /**
                     * DependencyType enum.
                     * @name fission.workflows.types.TaskDependencyParameters.DependencyType
                     * @enum {string}
                     * @property {number} DATA=0 DATA value
                     * @property {number} CONTROL=1 CONTROL value
                     * @property {number} DYNAMIC_OUTPUT=2 DYNAMIC_OUTPUT value
                     */
                    TaskDependencyParameters.DependencyType = (function() {
                        var valuesById = {}, values = Object.create(valuesById);
                        values[valuesById[0] = "DATA"] = 0;
                        values[valuesById[1] = "CONTROL"] = 1;
                        values[valuesById[2] = "DYNAMIC_OUTPUT"] = 2;
                        return values;
                    })();
    
                    return TaskDependencyParameters;
                })();
    
                types.TaskInvocation = (function() {
    
                    /**
                     * Properties of a TaskInvocation.
                     * @memberof fission.workflows.types
                     * @interface ITaskInvocation
                     * @property {fission.workflows.types.IObjectMetadata|null} [metadata] TaskInvocation metadata
                     * @property {fission.workflows.types.ITaskInvocationSpec|null} [spec] TaskInvocation spec
                     * @property {fission.workflows.types.ITaskInvocationStatus|null} [status] TaskInvocation status
                     */
    
                    /**
                     * Constructs a new TaskInvocation.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TaskInvocation.
                     * @implements ITaskInvocation
                     * @constructor
                     * @param {fission.workflows.types.ITaskInvocation=} [properties] Properties to set
                     */
                    function TaskInvocation(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TaskInvocation metadata.
                     * @member {fission.workflows.types.IObjectMetadata|null|undefined} metadata
                     * @memberof fission.workflows.types.TaskInvocation
                     * @instance
                     */
                    TaskInvocation.prototype.metadata = null;
    
                    /**
                     * TaskInvocation spec.
                     * @member {fission.workflows.types.ITaskInvocationSpec|null|undefined} spec
                     * @memberof fission.workflows.types.TaskInvocation
                     * @instance
                     */
                    TaskInvocation.prototype.spec = null;
    
                    /**
                     * TaskInvocation status.
                     * @member {fission.workflows.types.ITaskInvocationStatus|null|undefined} status
                     * @memberof fission.workflows.types.TaskInvocation
                     * @instance
                     */
                    TaskInvocation.prototype.status = null;
    
                    /**
                     * Creates a new TaskInvocation instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {fission.workflows.types.ITaskInvocation=} [properties] Properties to set
                     * @returns {fission.workflows.types.TaskInvocation} TaskInvocation instance
                     */
                    TaskInvocation.create = function create(properties) {
                        return new TaskInvocation(properties);
                    };
    
                    /**
                     * Encodes the specified TaskInvocation message. Does not implicitly {@link fission.workflows.types.TaskInvocation.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {fission.workflows.types.ITaskInvocation} message TaskInvocation message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskInvocation.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            $root.fission.workflows.types.ObjectMetadata.encode(message.metadata, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            $root.fission.workflows.types.TaskInvocationSpec.encode(message.spec, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.status != null && message.hasOwnProperty("status"))
                            $root.fission.workflows.types.TaskInvocationStatus.encode(message.status, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TaskInvocation message, length delimited. Does not implicitly {@link fission.workflows.types.TaskInvocation.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {fission.workflows.types.ITaskInvocation} message TaskInvocation message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskInvocation.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TaskInvocation message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TaskInvocation} TaskInvocation
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskInvocation.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TaskInvocation();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.metadata = $root.fission.workflows.types.ObjectMetadata.decode(reader, reader.uint32());
                                break;
                            case 2:
                                message.spec = $root.fission.workflows.types.TaskInvocationSpec.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.status = $root.fission.workflows.types.TaskInvocationStatus.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TaskInvocation message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TaskInvocation} TaskInvocation
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskInvocation.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TaskInvocation message.
                     * @function verify
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TaskInvocation.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.metadata != null && message.hasOwnProperty("metadata")) {
                            var error = $root.fission.workflows.types.ObjectMetadata.verify(message.metadata);
                            if (error)
                                return "metadata." + error;
                        }
                        if (message.spec != null && message.hasOwnProperty("spec")) {
                            var error = $root.fission.workflows.types.TaskInvocationSpec.verify(message.spec);
                            if (error)
                                return "spec." + error;
                        }
                        if (message.status != null && message.hasOwnProperty("status")) {
                            var error = $root.fission.workflows.types.TaskInvocationStatus.verify(message.status);
                            if (error)
                                return "status." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a TaskInvocation message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TaskInvocation} TaskInvocation
                     */
                    TaskInvocation.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TaskInvocation)
                            return object;
                        var message = new $root.fission.workflows.types.TaskInvocation();
                        if (object.metadata != null) {
                            if (typeof object.metadata !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocation.metadata: object expected");
                            message.metadata = $root.fission.workflows.types.ObjectMetadata.fromObject(object.metadata);
                        }
                        if (object.spec != null) {
                            if (typeof object.spec !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocation.spec: object expected");
                            message.spec = $root.fission.workflows.types.TaskInvocationSpec.fromObject(object.spec);
                        }
                        if (object.status != null) {
                            if (typeof object.status !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocation.status: object expected");
                            message.status = $root.fission.workflows.types.TaskInvocationStatus.fromObject(object.status);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TaskInvocation message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TaskInvocation
                     * @static
                     * @param {fission.workflows.types.TaskInvocation} message TaskInvocation
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TaskInvocation.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.metadata = null;
                            object.spec = null;
                            object.status = null;
                        }
                        if (message.metadata != null && message.hasOwnProperty("metadata"))
                            object.metadata = $root.fission.workflows.types.ObjectMetadata.toObject(message.metadata, options);
                        if (message.spec != null && message.hasOwnProperty("spec"))
                            object.spec = $root.fission.workflows.types.TaskInvocationSpec.toObject(message.spec, options);
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = $root.fission.workflows.types.TaskInvocationStatus.toObject(message.status, options);
                        return object;
                    };
    
                    /**
                     * Converts this TaskInvocation to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TaskInvocation
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TaskInvocation.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return TaskInvocation;
                })();
    
                types.TaskInvocationSpec = (function() {
    
                    /**
                     * Properties of a TaskInvocationSpec.
                     * @memberof fission.workflows.types
                     * @interface ITaskInvocationSpec
                     * @property {fission.workflows.types.IFnRef|null} [fnRef] TaskInvocationSpec fnRef
                     * @property {string|null} [taskId] TaskInvocationSpec taskId
                     * @property {Object.<string,fission.workflows.types.ITypedValue>|null} [inputs] TaskInvocationSpec inputs
                     * @property {string|null} [invocationId] TaskInvocationSpec invocationId
                     */
    
                    /**
                     * Constructs a new TaskInvocationSpec.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TaskInvocationSpec.
                     * @implements ITaskInvocationSpec
                     * @constructor
                     * @param {fission.workflows.types.ITaskInvocationSpec=} [properties] Properties to set
                     */
                    function TaskInvocationSpec(properties) {
                        this.inputs = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TaskInvocationSpec fnRef.
                     * @member {fission.workflows.types.IFnRef|null|undefined} fnRef
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @instance
                     */
                    TaskInvocationSpec.prototype.fnRef = null;
    
                    /**
                     * TaskInvocationSpec taskId.
                     * @member {string} taskId
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @instance
                     */
                    TaskInvocationSpec.prototype.taskId = "";
    
                    /**
                     * TaskInvocationSpec inputs.
                     * @member {Object.<string,fission.workflows.types.ITypedValue>} inputs
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @instance
                     */
                    TaskInvocationSpec.prototype.inputs = $util.emptyObject;
    
                    /**
                     * TaskInvocationSpec invocationId.
                     * @member {string} invocationId
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @instance
                     */
                    TaskInvocationSpec.prototype.invocationId = "";
    
                    /**
                     * Creates a new TaskInvocationSpec instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {fission.workflows.types.ITaskInvocationSpec=} [properties] Properties to set
                     * @returns {fission.workflows.types.TaskInvocationSpec} TaskInvocationSpec instance
                     */
                    TaskInvocationSpec.create = function create(properties) {
                        return new TaskInvocationSpec(properties);
                    };
    
                    /**
                     * Encodes the specified TaskInvocationSpec message. Does not implicitly {@link fission.workflows.types.TaskInvocationSpec.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {fission.workflows.types.ITaskInvocationSpec} message TaskInvocationSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskInvocationSpec.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.fnRef != null && message.hasOwnProperty("fnRef"))
                            $root.fission.workflows.types.FnRef.encode(message.fnRef, writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                        if (message.taskId != null && message.hasOwnProperty("taskId"))
                            writer.uint32(/* id 2, wireType 2 =*/18).string(message.taskId);
                        if (message.inputs != null && message.hasOwnProperty("inputs"))
                            for (var keys = Object.keys(message.inputs), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 3, wireType 2 =*/26).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TypedValue.encode(message.inputs[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        if (message.invocationId != null && message.hasOwnProperty("invocationId"))
                            writer.uint32(/* id 4, wireType 2 =*/34).string(message.invocationId);
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TaskInvocationSpec message, length delimited. Does not implicitly {@link fission.workflows.types.TaskInvocationSpec.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {fission.workflows.types.ITaskInvocationSpec} message TaskInvocationSpec message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskInvocationSpec.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TaskInvocationSpec message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TaskInvocationSpec} TaskInvocationSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskInvocationSpec.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TaskInvocationSpec(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.fnRef = $root.fission.workflows.types.FnRef.decode(reader, reader.uint32());
                                break;
                            case 2:
                                message.taskId = reader.string();
                                break;
                            case 3:
                                reader.skip().pos++;
                                if (message.inputs === $util.emptyObject)
                                    message.inputs = {};
                                key = reader.string();
                                reader.pos++;
                                message.inputs[key] = $root.fission.workflows.types.TypedValue.decode(reader, reader.uint32());
                                break;
                            case 4:
                                message.invocationId = reader.string();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TaskInvocationSpec message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TaskInvocationSpec} TaskInvocationSpec
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskInvocationSpec.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TaskInvocationSpec message.
                     * @function verify
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TaskInvocationSpec.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.fnRef != null && message.hasOwnProperty("fnRef")) {
                            var error = $root.fission.workflows.types.FnRef.verify(message.fnRef);
                            if (error)
                                return "fnRef." + error;
                        }
                        if (message.taskId != null && message.hasOwnProperty("taskId"))
                            if (!$util.isString(message.taskId))
                                return "taskId: string expected";
                        if (message.inputs != null && message.hasOwnProperty("inputs")) {
                            if (!$util.isObject(message.inputs))
                                return "inputs: object expected";
                            var key = Object.keys(message.inputs);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TypedValue.verify(message.inputs[key[i]]);
                                if (error)
                                    return "inputs." + error;
                            }
                        }
                        if (message.invocationId != null && message.hasOwnProperty("invocationId"))
                            if (!$util.isString(message.invocationId))
                                return "invocationId: string expected";
                        return null;
                    };
    
                    /**
                     * Creates a TaskInvocationSpec message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TaskInvocationSpec} TaskInvocationSpec
                     */
                    TaskInvocationSpec.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TaskInvocationSpec)
                            return object;
                        var message = new $root.fission.workflows.types.TaskInvocationSpec();
                        if (object.fnRef != null) {
                            if (typeof object.fnRef !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocationSpec.fnRef: object expected");
                            message.fnRef = $root.fission.workflows.types.FnRef.fromObject(object.fnRef);
                        }
                        if (object.taskId != null)
                            message.taskId = String(object.taskId);
                        if (object.inputs) {
                            if (typeof object.inputs !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocationSpec.inputs: object expected");
                            message.inputs = {};
                            for (var keys = Object.keys(object.inputs), i = 0; i < keys.length; ++i) {
                                if (typeof object.inputs[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.TaskInvocationSpec.inputs: object expected");
                                message.inputs[keys[i]] = $root.fission.workflows.types.TypedValue.fromObject(object.inputs[keys[i]]);
                            }
                        }
                        if (object.invocationId != null)
                            message.invocationId = String(object.invocationId);
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TaskInvocationSpec message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @static
                     * @param {fission.workflows.types.TaskInvocationSpec} message TaskInvocationSpec
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TaskInvocationSpec.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults)
                            object.inputs = {};
                        if (options.defaults) {
                            object.fnRef = null;
                            object.taskId = "";
                            object.invocationId = "";
                        }
                        if (message.fnRef != null && message.hasOwnProperty("fnRef"))
                            object.fnRef = $root.fission.workflows.types.FnRef.toObject(message.fnRef, options);
                        if (message.taskId != null && message.hasOwnProperty("taskId"))
                            object.taskId = message.taskId;
                        var keys2;
                        if (message.inputs && (keys2 = Object.keys(message.inputs)).length) {
                            object.inputs = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.inputs[keys2[j]] = $root.fission.workflows.types.TypedValue.toObject(message.inputs[keys2[j]], options);
                        }
                        if (message.invocationId != null && message.hasOwnProperty("invocationId"))
                            object.invocationId = message.invocationId;
                        return object;
                    };
    
                    /**
                     * Converts this TaskInvocationSpec to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TaskInvocationSpec
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TaskInvocationSpec.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return TaskInvocationSpec;
                })();
    
                types.TaskInvocationStatus = (function() {
    
                    /**
                     * Properties of a TaskInvocationStatus.
                     * @memberof fission.workflows.types
                     * @interface ITaskInvocationStatus
                     * @property {fission.workflows.types.TaskInvocationStatus.Status|null} [status] TaskInvocationStatus status
                     * @property {google.protobuf.ITimestamp|null} [updatedAt] TaskInvocationStatus updatedAt
                     * @property {fission.workflows.types.ITypedValue|null} [output] TaskInvocationStatus output
                     * @property {fission.workflows.types.IError|null} [error] TaskInvocationStatus error
                     */
    
                    /**
                     * Constructs a new TaskInvocationStatus.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TaskInvocationStatus.
                     * @implements ITaskInvocationStatus
                     * @constructor
                     * @param {fission.workflows.types.ITaskInvocationStatus=} [properties] Properties to set
                     */
                    function TaskInvocationStatus(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TaskInvocationStatus status.
                     * @member {fission.workflows.types.TaskInvocationStatus.Status} status
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @instance
                     */
                    TaskInvocationStatus.prototype.status = 0;
    
                    /**
                     * TaskInvocationStatus updatedAt.
                     * @member {google.protobuf.ITimestamp|null|undefined} updatedAt
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @instance
                     */
                    TaskInvocationStatus.prototype.updatedAt = null;
    
                    /**
                     * TaskInvocationStatus output.
                     * @member {fission.workflows.types.ITypedValue|null|undefined} output
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @instance
                     */
                    TaskInvocationStatus.prototype.output = null;
    
                    /**
                     * TaskInvocationStatus error.
                     * @member {fission.workflows.types.IError|null|undefined} error
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @instance
                     */
                    TaskInvocationStatus.prototype.error = null;
    
                    /**
                     * Creates a new TaskInvocationStatus instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {fission.workflows.types.ITaskInvocationStatus=} [properties] Properties to set
                     * @returns {fission.workflows.types.TaskInvocationStatus} TaskInvocationStatus instance
                     */
                    TaskInvocationStatus.create = function create(properties) {
                        return new TaskInvocationStatus(properties);
                    };
    
                    /**
                     * Encodes the specified TaskInvocationStatus message. Does not implicitly {@link fission.workflows.types.TaskInvocationStatus.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {fission.workflows.types.ITaskInvocationStatus} message TaskInvocationStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskInvocationStatus.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.status != null && message.hasOwnProperty("status"))
                            writer.uint32(/* id 1, wireType 0 =*/8).int32(message.status);
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            $root.google.protobuf.Timestamp.encode(message.updatedAt, writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
                        if (message.output != null && message.hasOwnProperty("output"))
                            $root.fission.workflows.types.TypedValue.encode(message.output, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                        if (message.error != null && message.hasOwnProperty("error"))
                            $root.fission.workflows.types.Error.encode(message.error, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TaskInvocationStatus message, length delimited. Does not implicitly {@link fission.workflows.types.TaskInvocationStatus.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {fission.workflows.types.ITaskInvocationStatus} message TaskInvocationStatus message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TaskInvocationStatus.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TaskInvocationStatus message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TaskInvocationStatus} TaskInvocationStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskInvocationStatus.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TaskInvocationStatus();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.status = reader.int32();
                                break;
                            case 2:
                                message.updatedAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                                break;
                            case 3:
                                message.output = $root.fission.workflows.types.TypedValue.decode(reader, reader.uint32());
                                break;
                            case 4:
                                message.error = $root.fission.workflows.types.Error.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TaskInvocationStatus message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TaskInvocationStatus} TaskInvocationStatus
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TaskInvocationStatus.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TaskInvocationStatus message.
                     * @function verify
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TaskInvocationStatus.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.status != null && message.hasOwnProperty("status"))
                            switch (message.status) {
                            default:
                                return "status: enum value expected";
                            case 0:
                            case 1:
                            case 2:
                            case 3:
                            case 4:
                            case 5:
                            case 6:
                                break;
                            }
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt")) {
                            var error = $root.google.protobuf.Timestamp.verify(message.updatedAt);
                            if (error)
                                return "updatedAt." + error;
                        }
                        if (message.output != null && message.hasOwnProperty("output")) {
                            var error = $root.fission.workflows.types.TypedValue.verify(message.output);
                            if (error)
                                return "output." + error;
                        }
                        if (message.error != null && message.hasOwnProperty("error")) {
                            var error = $root.fission.workflows.types.Error.verify(message.error);
                            if (error)
                                return "error." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates a TaskInvocationStatus message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TaskInvocationStatus} TaskInvocationStatus
                     */
                    TaskInvocationStatus.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TaskInvocationStatus)
                            return object;
                        var message = new $root.fission.workflows.types.TaskInvocationStatus();
                        switch (object.status) {
                        case "UNKNOWN":
                        case 0:
                            message.status = 0;
                            break;
                        case "SCHEDULED":
                        case 1:
                            message.status = 1;
                            break;
                        case "IN_PROGRESS":
                        case 2:
                            message.status = 2;
                            break;
                        case "SUCCEEDED":
                        case 3:
                            message.status = 3;
                            break;
                        case "FAILED":
                        case 4:
                            message.status = 4;
                            break;
                        case "ABORTED":
                        case 5:
                            message.status = 5;
                            break;
                        case "SKIPPED":
                        case 6:
                            message.status = 6;
                            break;
                        }
                        if (object.updatedAt != null) {
                            if (typeof object.updatedAt !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocationStatus.updatedAt: object expected");
                            message.updatedAt = $root.google.protobuf.Timestamp.fromObject(object.updatedAt);
                        }
                        if (object.output != null) {
                            if (typeof object.output !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocationStatus.output: object expected");
                            message.output = $root.fission.workflows.types.TypedValue.fromObject(object.output);
                        }
                        if (object.error != null) {
                            if (typeof object.error !== "object")
                                throw TypeError(".fission.workflows.types.TaskInvocationStatus.error: object expected");
                            message.error = $root.fission.workflows.types.Error.fromObject(object.error);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TaskInvocationStatus message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @static
                     * @param {fission.workflows.types.TaskInvocationStatus} message TaskInvocationStatus
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TaskInvocationStatus.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.status = options.enums === String ? "UNKNOWN" : 0;
                            object.updatedAt = null;
                            object.output = null;
                            object.error = null;
                        }
                        if (message.status != null && message.hasOwnProperty("status"))
                            object.status = options.enums === String ? $root.fission.workflows.types.TaskInvocationStatus.Status[message.status] : message.status;
                        if (message.updatedAt != null && message.hasOwnProperty("updatedAt"))
                            object.updatedAt = $root.google.protobuf.Timestamp.toObject(message.updatedAt, options);
                        if (message.output != null && message.hasOwnProperty("output"))
                            object.output = $root.fission.workflows.types.TypedValue.toObject(message.output, options);
                        if (message.error != null && message.hasOwnProperty("error"))
                            object.error = $root.fission.workflows.types.Error.toObject(message.error, options);
                        return object;
                    };
    
                    /**
                     * Converts this TaskInvocationStatus to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TaskInvocationStatus
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TaskInvocationStatus.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    /**
                     * Status enum.
                     * @name fission.workflows.types.TaskInvocationStatus.Status
                     * @enum {string}
                     * @property {number} UNKNOWN=0 UNKNOWN value
                     * @property {number} SCHEDULED=1 SCHEDULED value
                     * @property {number} IN_PROGRESS=2 IN_PROGRESS value
                     * @property {number} SUCCEEDED=3 SUCCEEDED value
                     * @property {number} FAILED=4 FAILED value
                     * @property {number} ABORTED=5 ABORTED value
                     * @property {number} SKIPPED=6 SKIPPED value
                     */
                    TaskInvocationStatus.Status = (function() {
                        var valuesById = {}, values = Object.create(valuesById);
                        values[valuesById[0] = "UNKNOWN"] = 0;
                        values[valuesById[1] = "SCHEDULED"] = 1;
                        values[valuesById[2] = "IN_PROGRESS"] = 2;
                        values[valuesById[3] = "SUCCEEDED"] = 3;
                        values[valuesById[4] = "FAILED"] = 4;
                        values[valuesById[5] = "ABORTED"] = 5;
                        values[valuesById[6] = "SKIPPED"] = 6;
                        return values;
                    })();
    
                    return TaskInvocationStatus;
                })();
    
                types.ObjectMetadata = (function() {
    
                    /**
                     * Properties of an ObjectMetadata.
                     * @memberof fission.workflows.types
                     * @interface IObjectMetadata
                     * @property {string|null} [id] ObjectMetadata id
                     * @property {google.protobuf.ITimestamp|null} [createdAt] ObjectMetadata createdAt
                     */
    
                    /**
                     * Constructs a new ObjectMetadata.
                     * @memberof fission.workflows.types
                     * @classdesc Represents an ObjectMetadata.
                     * @implements IObjectMetadata
                     * @constructor
                     * @param {fission.workflows.types.IObjectMetadata=} [properties] Properties to set
                     */
                    function ObjectMetadata(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * ObjectMetadata id.
                     * @member {string} id
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @instance
                     */
                    ObjectMetadata.prototype.id = "";
    
                    /**
                     * ObjectMetadata createdAt.
                     * @member {google.protobuf.ITimestamp|null|undefined} createdAt
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @instance
                     */
                    ObjectMetadata.prototype.createdAt = null;
    
                    /**
                     * Creates a new ObjectMetadata instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {fission.workflows.types.IObjectMetadata=} [properties] Properties to set
                     * @returns {fission.workflows.types.ObjectMetadata} ObjectMetadata instance
                     */
                    ObjectMetadata.create = function create(properties) {
                        return new ObjectMetadata(properties);
                    };
    
                    /**
                     * Encodes the specified ObjectMetadata message. Does not implicitly {@link fission.workflows.types.ObjectMetadata.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {fission.workflows.types.IObjectMetadata} message ObjectMetadata message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    ObjectMetadata.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.id != null && message.hasOwnProperty("id"))
                            writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
                        if (message.createdAt != null && message.hasOwnProperty("createdAt"))
                            $root.google.protobuf.Timestamp.encode(message.createdAt, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified ObjectMetadata message, length delimited. Does not implicitly {@link fission.workflows.types.ObjectMetadata.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {fission.workflows.types.IObjectMetadata} message ObjectMetadata message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    ObjectMetadata.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes an ObjectMetadata message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.ObjectMetadata} ObjectMetadata
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    ObjectMetadata.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.ObjectMetadata();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.id = reader.string();
                                break;
                            case 3:
                                message.createdAt = $root.google.protobuf.Timestamp.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes an ObjectMetadata message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.ObjectMetadata} ObjectMetadata
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    ObjectMetadata.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies an ObjectMetadata message.
                     * @function verify
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    ObjectMetadata.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.id != null && message.hasOwnProperty("id"))
                            if (!$util.isString(message.id))
                                return "id: string expected";
                        if (message.createdAt != null && message.hasOwnProperty("createdAt")) {
                            var error = $root.google.protobuf.Timestamp.verify(message.createdAt);
                            if (error)
                                return "createdAt." + error;
                        }
                        return null;
                    };
    
                    /**
                     * Creates an ObjectMetadata message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.ObjectMetadata} ObjectMetadata
                     */
                    ObjectMetadata.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.ObjectMetadata)
                            return object;
                        var message = new $root.fission.workflows.types.ObjectMetadata();
                        if (object.id != null)
                            message.id = String(object.id);
                        if (object.createdAt != null) {
                            if (typeof object.createdAt !== "object")
                                throw TypeError(".fission.workflows.types.ObjectMetadata.createdAt: object expected");
                            message.createdAt = $root.google.protobuf.Timestamp.fromObject(object.createdAt);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from an ObjectMetadata message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @static
                     * @param {fission.workflows.types.ObjectMetadata} message ObjectMetadata
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    ObjectMetadata.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.id = "";
                            object.createdAt = null;
                        }
                        if (message.id != null && message.hasOwnProperty("id"))
                            object.id = message.id;
                        if (message.createdAt != null && message.hasOwnProperty("createdAt"))
                            object.createdAt = $root.google.protobuf.Timestamp.toObject(message.createdAt, options);
                        return object;
                    };
    
                    /**
                     * Converts this ObjectMetadata to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.ObjectMetadata
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    ObjectMetadata.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return ObjectMetadata;
                })();
    
                types.TypedValue = (function() {
    
                    /**
                     * Properties of a TypedValue.
                     * @memberof fission.workflows.types
                     * @interface ITypedValue
                     * @property {string|null} [type] TypedValue type
                     * @property {Uint8Array|null} [value] TypedValue value
                     * @property {Object.<string,string>|null} [labels] TypedValue labels
                     */
    
                    /**
                     * Constructs a new TypedValue.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TypedValue.
                     * @implements ITypedValue
                     * @constructor
                     * @param {fission.workflows.types.ITypedValue=} [properties] Properties to set
                     */
                    function TypedValue(properties) {
                        this.labels = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TypedValue type.
                     * @member {string} type
                     * @memberof fission.workflows.types.TypedValue
                     * @instance
                     */
                    TypedValue.prototype.type = "";
    
                    /**
                     * TypedValue value.
                     * @member {Uint8Array} value
                     * @memberof fission.workflows.types.TypedValue
                     * @instance
                     */
                    TypedValue.prototype.value = $util.newBuffer([]);
    
                    /**
                     * TypedValue labels.
                     * @member {Object.<string,string>} labels
                     * @memberof fission.workflows.types.TypedValue
                     * @instance
                     */
                    TypedValue.prototype.labels = $util.emptyObject;
    
                    /**
                     * Creates a new TypedValue instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {fission.workflows.types.ITypedValue=} [properties] Properties to set
                     * @returns {fission.workflows.types.TypedValue} TypedValue instance
                     */
                    TypedValue.create = function create(properties) {
                        return new TypedValue(properties);
                    };
    
                    /**
                     * Encodes the specified TypedValue message. Does not implicitly {@link fission.workflows.types.TypedValue.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {fission.workflows.types.ITypedValue} message TypedValue message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TypedValue.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.type != null && message.hasOwnProperty("type"))
                            writer.uint32(/* id 1, wireType 2 =*/10).string(message.type);
                        if (message.value != null && message.hasOwnProperty("value"))
                            writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.value);
                        if (message.labels != null && message.hasOwnProperty("labels"))
                            for (var keys = Object.keys(message.labels), i = 0; i < keys.length; ++i)
                                writer.uint32(/* id 3, wireType 2 =*/26).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]).uint32(/* id 2, wireType 2 =*/18).string(message.labels[keys[i]]).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TypedValue message, length delimited. Does not implicitly {@link fission.workflows.types.TypedValue.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {fission.workflows.types.ITypedValue} message TypedValue message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TypedValue.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TypedValue message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TypedValue} TypedValue
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TypedValue.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TypedValue(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                message.type = reader.string();
                                break;
                            case 2:
                                message.value = reader.bytes();
                                break;
                            case 3:
                                reader.skip().pos++;
                                if (message.labels === $util.emptyObject)
                                    message.labels = {};
                                key = reader.string();
                                reader.pos++;
                                message.labels[key] = reader.string();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TypedValue message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TypedValue} TypedValue
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TypedValue.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TypedValue message.
                     * @function verify
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TypedValue.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.type != null && message.hasOwnProperty("type"))
                            if (!$util.isString(message.type))
                                return "type: string expected";
                        if (message.value != null && message.hasOwnProperty("value"))
                            if (!(message.value && typeof message.value.length === "number" || $util.isString(message.value)))
                                return "value: buffer expected";
                        if (message.labels != null && message.hasOwnProperty("labels")) {
                            if (!$util.isObject(message.labels))
                                return "labels: object expected";
                            var key = Object.keys(message.labels);
                            for (var i = 0; i < key.length; ++i)
                                if (!$util.isString(message.labels[key[i]]))
                                    return "labels: string{k:string} expected";
                        }
                        return null;
                    };
    
                    /**
                     * Creates a TypedValue message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TypedValue} TypedValue
                     */
                    TypedValue.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TypedValue)
                            return object;
                        var message = new $root.fission.workflows.types.TypedValue();
                        if (object.type != null)
                            message.type = String(object.type);
                        if (object.value != null)
                            if (typeof object.value === "string")
                                $util.base64.decode(object.value, message.value = $util.newBuffer($util.base64.length(object.value)), 0);
                            else if (object.value.length)
                                message.value = object.value;
                        if (object.labels) {
                            if (typeof object.labels !== "object")
                                throw TypeError(".fission.workflows.types.TypedValue.labels: object expected");
                            message.labels = {};
                            for (var keys = Object.keys(object.labels), i = 0; i < keys.length; ++i)
                                message.labels[keys[i]] = String(object.labels[keys[i]]);
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TypedValue message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TypedValue
                     * @static
                     * @param {fission.workflows.types.TypedValue} message TypedValue
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TypedValue.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults)
                            object.labels = {};
                        if (options.defaults) {
                            object.type = "";
                            if (options.bytes === String)
                                object.value = "";
                            else {
                                object.value = [];
                                if (options.bytes !== Array)
                                    object.value = $util.newBuffer(object.value);
                            }
                        }
                        if (message.type != null && message.hasOwnProperty("type"))
                            object.type = message.type;
                        if (message.value != null && message.hasOwnProperty("value"))
                            object.value = options.bytes === String ? $util.base64.encode(message.value, 0, message.value.length) : options.bytes === Array ? Array.prototype.slice.call(message.value) : message.value;
                        var keys2;
                        if (message.labels && (keys2 = Object.keys(message.labels)).length) {
                            object.labels = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.labels[keys2[j]] = message.labels[keys2[j]];
                        }
                        return object;
                    };
    
                    /**
                     * Converts this TypedValue to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TypedValue
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TypedValue.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return TypedValue;
                })();
    
                types.Error = (function() {
    
                    /**
                     * Properties of an Error.
                     * @memberof fission.workflows.types
                     * @interface IError
                     * @property {string|null} [message] Error message
                     */
    
                    /**
                     * Constructs a new Error.
                     * @memberof fission.workflows.types
                     * @classdesc Represents an Error.
                     * @implements IError
                     * @constructor
                     * @param {fission.workflows.types.IError=} [properties] Properties to set
                     */
                    function Error(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * Error message.
                     * @member {string} message
                     * @memberof fission.workflows.types.Error
                     * @instance
                     */
                    Error.prototype.message = "";
    
                    /**
                     * Creates a new Error instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {fission.workflows.types.IError=} [properties] Properties to set
                     * @returns {fission.workflows.types.Error} Error instance
                     */
                    Error.create = function create(properties) {
                        return new Error(properties);
                    };
    
                    /**
                     * Encodes the specified Error message. Does not implicitly {@link fission.workflows.types.Error.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {fission.workflows.types.IError} message Error message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    Error.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.message != null && message.hasOwnProperty("message"))
                            writer.uint32(/* id 2, wireType 2 =*/18).string(message.message);
                        return writer;
                    };
    
                    /**
                     * Encodes the specified Error message, length delimited. Does not implicitly {@link fission.workflows.types.Error.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {fission.workflows.types.IError} message Error message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    Error.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes an Error message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.Error} Error
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    Error.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.Error();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 2:
                                message.message = reader.string();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes an Error message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.Error} Error
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    Error.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies an Error message.
                     * @function verify
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    Error.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.message != null && message.hasOwnProperty("message"))
                            if (!$util.isString(message.message))
                                return "message: string expected";
                        return null;
                    };
    
                    /**
                     * Creates an Error message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.Error} Error
                     */
                    Error.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.Error)
                            return object;
                        var message = new $root.fission.workflows.types.Error();
                        if (object.message != null)
                            message.message = String(object.message);
                        return message;
                    };
    
                    /**
                     * Creates a plain object from an Error message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.Error
                     * @static
                     * @param {fission.workflows.types.Error} message Error
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    Error.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults)
                            object.message = "";
                        if (message.message != null && message.hasOwnProperty("message"))
                            object.message = message.message;
                        return object;
                    };
    
                    /**
                     * Converts this Error to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.Error
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    Error.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return Error;
                })();
    
                types.FnRef = (function() {
    
                    /**
                     * Properties of a FnRef.
                     * @memberof fission.workflows.types
                     * @interface IFnRef
                     * @property {string|null} [runtime] FnRef runtime
                     * @property {string|null} [namespace] FnRef namespace
                     * @property {string|null} [ID] FnRef ID
                     */
    
                    /**
                     * Constructs a new FnRef.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a FnRef.
                     * @implements IFnRef
                     * @constructor
                     * @param {fission.workflows.types.IFnRef=} [properties] Properties to set
                     */
                    function FnRef(properties) {
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * FnRef runtime.
                     * @member {string} runtime
                     * @memberof fission.workflows.types.FnRef
                     * @instance
                     */
                    FnRef.prototype.runtime = "";
    
                    /**
                     * FnRef namespace.
                     * @member {string} namespace
                     * @memberof fission.workflows.types.FnRef
                     * @instance
                     */
                    FnRef.prototype.namespace = "";
    
                    /**
                     * FnRef ID.
                     * @member {string} ID
                     * @memberof fission.workflows.types.FnRef
                     * @instance
                     */
                    FnRef.prototype.ID = "";
    
                    /**
                     * Creates a new FnRef instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {fission.workflows.types.IFnRef=} [properties] Properties to set
                     * @returns {fission.workflows.types.FnRef} FnRef instance
                     */
                    FnRef.create = function create(properties) {
                        return new FnRef(properties);
                    };
    
                    /**
                     * Encodes the specified FnRef message. Does not implicitly {@link fission.workflows.types.FnRef.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {fission.workflows.types.IFnRef} message FnRef message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    FnRef.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.runtime != null && message.hasOwnProperty("runtime"))
                            writer.uint32(/* id 2, wireType 2 =*/18).string(message.runtime);
                        if (message.namespace != null && message.hasOwnProperty("namespace"))
                            writer.uint32(/* id 3, wireType 2 =*/26).string(message.namespace);
                        if (message.ID != null && message.hasOwnProperty("ID"))
                            writer.uint32(/* id 4, wireType 2 =*/34).string(message.ID);
                        return writer;
                    };
    
                    /**
                     * Encodes the specified FnRef message, length delimited. Does not implicitly {@link fission.workflows.types.FnRef.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {fission.workflows.types.IFnRef} message FnRef message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    FnRef.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a FnRef message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.FnRef} FnRef
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    FnRef.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.FnRef();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 2:
                                message.runtime = reader.string();
                                break;
                            case 3:
                                message.namespace = reader.string();
                                break;
                            case 4:
                                message.ID = reader.string();
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a FnRef message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.FnRef} FnRef
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    FnRef.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a FnRef message.
                     * @function verify
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    FnRef.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.runtime != null && message.hasOwnProperty("runtime"))
                            if (!$util.isString(message.runtime))
                                return "runtime: string expected";
                        if (message.namespace != null && message.hasOwnProperty("namespace"))
                            if (!$util.isString(message.namespace))
                                return "namespace: string expected";
                        if (message.ID != null && message.hasOwnProperty("ID"))
                            if (!$util.isString(message.ID))
                                return "ID: string expected";
                        return null;
                    };
    
                    /**
                     * Creates a FnRef message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.FnRef} FnRef
                     */
                    FnRef.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.FnRef)
                            return object;
                        var message = new $root.fission.workflows.types.FnRef();
                        if (object.runtime != null)
                            message.runtime = String(object.runtime);
                        if (object.namespace != null)
                            message.namespace = String(object.namespace);
                        if (object.ID != null)
                            message.ID = String(object.ID);
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a FnRef message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.FnRef
                     * @static
                     * @param {fission.workflows.types.FnRef} message FnRef
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    FnRef.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.defaults) {
                            object.runtime = "";
                            object.namespace = "";
                            object.ID = "";
                        }
                        if (message.runtime != null && message.hasOwnProperty("runtime"))
                            object.runtime = message.runtime;
                        if (message.namespace != null && message.hasOwnProperty("namespace"))
                            object.namespace = message.namespace;
                        if (message.ID != null && message.hasOwnProperty("ID"))
                            object.ID = message.ID;
                        return object;
                    };
    
                    /**
                     * Converts this FnRef to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.FnRef
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    FnRef.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return FnRef;
                })();
    
                types.TypedValueMap = (function() {
    
                    /**
                     * Properties of a TypedValueMap.
                     * @memberof fission.workflows.types
                     * @interface ITypedValueMap
                     * @property {Object.<string,fission.workflows.types.ITypedValue>|null} [Value] TypedValueMap Value
                     */
    
                    /**
                     * Constructs a new TypedValueMap.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TypedValueMap.
                     * @implements ITypedValueMap
                     * @constructor
                     * @param {fission.workflows.types.ITypedValueMap=} [properties] Properties to set
                     */
                    function TypedValueMap(properties) {
                        this.Value = {};
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TypedValueMap Value.
                     * @member {Object.<string,fission.workflows.types.ITypedValue>} Value
                     * @memberof fission.workflows.types.TypedValueMap
                     * @instance
                     */
                    TypedValueMap.prototype.Value = $util.emptyObject;
    
                    /**
                     * Creates a new TypedValueMap instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {fission.workflows.types.ITypedValueMap=} [properties] Properties to set
                     * @returns {fission.workflows.types.TypedValueMap} TypedValueMap instance
                     */
                    TypedValueMap.create = function create(properties) {
                        return new TypedValueMap(properties);
                    };
    
                    /**
                     * Encodes the specified TypedValueMap message. Does not implicitly {@link fission.workflows.types.TypedValueMap.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {fission.workflows.types.ITypedValueMap} message TypedValueMap message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TypedValueMap.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.Value != null && message.hasOwnProperty("Value"))
                            for (var keys = Object.keys(message.Value), i = 0; i < keys.length; ++i) {
                                writer.uint32(/* id 1, wireType 2 =*/10).fork().uint32(/* id 1, wireType 2 =*/10).string(keys[i]);
                                $root.fission.workflows.types.TypedValue.encode(message.Value[keys[i]], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim().ldelim();
                            }
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TypedValueMap message, length delimited. Does not implicitly {@link fission.workflows.types.TypedValueMap.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {fission.workflows.types.ITypedValueMap} message TypedValueMap message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TypedValueMap.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TypedValueMap message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TypedValueMap} TypedValueMap
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TypedValueMap.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TypedValueMap(), key;
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                reader.skip().pos++;
                                if (message.Value === $util.emptyObject)
                                    message.Value = {};
                                key = reader.string();
                                reader.pos++;
                                message.Value[key] = $root.fission.workflows.types.TypedValue.decode(reader, reader.uint32());
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TypedValueMap message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TypedValueMap} TypedValueMap
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TypedValueMap.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TypedValueMap message.
                     * @function verify
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TypedValueMap.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.Value != null && message.hasOwnProperty("Value")) {
                            if (!$util.isObject(message.Value))
                                return "Value: object expected";
                            var key = Object.keys(message.Value);
                            for (var i = 0; i < key.length; ++i) {
                                var error = $root.fission.workflows.types.TypedValue.verify(message.Value[key[i]]);
                                if (error)
                                    return "Value." + error;
                            }
                        }
                        return null;
                    };
    
                    /**
                     * Creates a TypedValueMap message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TypedValueMap} TypedValueMap
                     */
                    TypedValueMap.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TypedValueMap)
                            return object;
                        var message = new $root.fission.workflows.types.TypedValueMap();
                        if (object.Value) {
                            if (typeof object.Value !== "object")
                                throw TypeError(".fission.workflows.types.TypedValueMap.Value: object expected");
                            message.Value = {};
                            for (var keys = Object.keys(object.Value), i = 0; i < keys.length; ++i) {
                                if (typeof object.Value[keys[i]] !== "object")
                                    throw TypeError(".fission.workflows.types.TypedValueMap.Value: object expected");
                                message.Value[keys[i]] = $root.fission.workflows.types.TypedValue.fromObject(object.Value[keys[i]]);
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TypedValueMap message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TypedValueMap
                     * @static
                     * @param {fission.workflows.types.TypedValueMap} message TypedValueMap
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TypedValueMap.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.objects || options.defaults)
                            object.Value = {};
                        var keys2;
                        if (message.Value && (keys2 = Object.keys(message.Value)).length) {
                            object.Value = {};
                            for (var j = 0; j < keys2.length; ++j)
                                object.Value[keys2[j]] = $root.fission.workflows.types.TypedValue.toObject(message.Value[keys2[j]], options);
                        }
                        return object;
                    };
    
                    /**
                     * Converts this TypedValueMap to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TypedValueMap
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TypedValueMap.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return TypedValueMap;
                })();
    
                types.TypedValueList = (function() {
    
                    /**
                     * Properties of a TypedValueList.
                     * @memberof fission.workflows.types
                     * @interface ITypedValueList
                     * @property {Array.<fission.workflows.types.ITypedValue>|null} [Value] TypedValueList Value
                     */
    
                    /**
                     * Constructs a new TypedValueList.
                     * @memberof fission.workflows.types
                     * @classdesc Represents a TypedValueList.
                     * @implements ITypedValueList
                     * @constructor
                     * @param {fission.workflows.types.ITypedValueList=} [properties] Properties to set
                     */
                    function TypedValueList(properties) {
                        this.Value = [];
                        if (properties)
                            for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                                if (properties[keys[i]] != null)
                                    this[keys[i]] = properties[keys[i]];
                    }
    
                    /**
                     * TypedValueList Value.
                     * @member {Array.<fission.workflows.types.ITypedValue>} Value
                     * @memberof fission.workflows.types.TypedValueList
                     * @instance
                     */
                    TypedValueList.prototype.Value = $util.emptyArray;
    
                    /**
                     * Creates a new TypedValueList instance using the specified properties.
                     * @function create
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {fission.workflows.types.ITypedValueList=} [properties] Properties to set
                     * @returns {fission.workflows.types.TypedValueList} TypedValueList instance
                     */
                    TypedValueList.create = function create(properties) {
                        return new TypedValueList(properties);
                    };
    
                    /**
                     * Encodes the specified TypedValueList message. Does not implicitly {@link fission.workflows.types.TypedValueList.verify|verify} messages.
                     * @function encode
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {fission.workflows.types.ITypedValueList} message TypedValueList message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TypedValueList.encode = function encode(message, writer) {
                        if (!writer)
                            writer = $Writer.create();
                        if (message.Value != null && message.Value.length)
                            for (var i = 0; i < message.Value.length; ++i)
                                $root.fission.workflows.types.TypedValue.encode(message.Value[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
                        return writer;
                    };
    
                    /**
                     * Encodes the specified TypedValueList message, length delimited. Does not implicitly {@link fission.workflows.types.TypedValueList.verify|verify} messages.
                     * @function encodeDelimited
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {fission.workflows.types.ITypedValueList} message TypedValueList message or plain object to encode
                     * @param {$protobuf.Writer} [writer] Writer to encode to
                     * @returns {$protobuf.Writer} Writer
                     */
                    TypedValueList.encodeDelimited = function encodeDelimited(message, writer) {
                        return this.encode(message, writer).ldelim();
                    };
    
                    /**
                     * Decodes a TypedValueList message from the specified reader or buffer.
                     * @function decode
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @param {number} [length] Message length if known beforehand
                     * @returns {fission.workflows.types.TypedValueList} TypedValueList
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TypedValueList.decode = function decode(reader, length) {
                        if (!(reader instanceof $Reader))
                            reader = $Reader.create(reader);
                        var end = length === undefined ? reader.len : reader.pos + length, message = new $root.fission.workflows.types.TypedValueList();
                        while (reader.pos < end) {
                            var tag = reader.uint32();
                            switch (tag >>> 3) {
                            case 1:
                                if (!(message.Value && message.Value.length))
                                    message.Value = [];
                                message.Value.push($root.fission.workflows.types.TypedValue.decode(reader, reader.uint32()));
                                break;
                            default:
                                reader.skipType(tag & 7);
                                break;
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Decodes a TypedValueList message from the specified reader or buffer, length delimited.
                     * @function decodeDelimited
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                     * @returns {fission.workflows.types.TypedValueList} TypedValueList
                     * @throws {Error} If the payload is not a reader or valid buffer
                     * @throws {$protobuf.util.ProtocolError} If required fields are missing
                     */
                    TypedValueList.decodeDelimited = function decodeDelimited(reader) {
                        if (!(reader instanceof $Reader))
                            reader = new $Reader(reader);
                        return this.decode(reader, reader.uint32());
                    };
    
                    /**
                     * Verifies a TypedValueList message.
                     * @function verify
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {Object.<string,*>} message Plain object to verify
                     * @returns {string|null} `null` if valid, otherwise the reason why it is not
                     */
                    TypedValueList.verify = function verify(message) {
                        if (typeof message !== "object" || message === null)
                            return "object expected";
                        if (message.Value != null && message.hasOwnProperty("Value")) {
                            if (!Array.isArray(message.Value))
                                return "Value: array expected";
                            for (var i = 0; i < message.Value.length; ++i) {
                                var error = $root.fission.workflows.types.TypedValue.verify(message.Value[i]);
                                if (error)
                                    return "Value." + error;
                            }
                        }
                        return null;
                    };
    
                    /**
                     * Creates a TypedValueList message from a plain object. Also converts values to their respective internal types.
                     * @function fromObject
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {Object.<string,*>} object Plain object
                     * @returns {fission.workflows.types.TypedValueList} TypedValueList
                     */
                    TypedValueList.fromObject = function fromObject(object) {
                        if (object instanceof $root.fission.workflows.types.TypedValueList)
                            return object;
                        var message = new $root.fission.workflows.types.TypedValueList();
                        if (object.Value) {
                            if (!Array.isArray(object.Value))
                                throw TypeError(".fission.workflows.types.TypedValueList.Value: array expected");
                            message.Value = [];
                            for (var i = 0; i < object.Value.length; ++i) {
                                if (typeof object.Value[i] !== "object")
                                    throw TypeError(".fission.workflows.types.TypedValueList.Value: object expected");
                                message.Value[i] = $root.fission.workflows.types.TypedValue.fromObject(object.Value[i]);
                            }
                        }
                        return message;
                    };
    
                    /**
                     * Creates a plain object from a TypedValueList message. Also converts values to other types if specified.
                     * @function toObject
                     * @memberof fission.workflows.types.TypedValueList
                     * @static
                     * @param {fission.workflows.types.TypedValueList} message TypedValueList
                     * @param {$protobuf.IConversionOptions} [options] Conversion options
                     * @returns {Object.<string,*>} Plain object
                     */
                    TypedValueList.toObject = function toObject(message, options) {
                        if (!options)
                            options = {};
                        var object = {};
                        if (options.arrays || options.defaults)
                            object.Value = [];
                        if (message.Value && message.Value.length) {
                            object.Value = [];
                            for (var j = 0; j < message.Value.length; ++j)
                                object.Value[j] = $root.fission.workflows.types.TypedValue.toObject(message.Value[j], options);
                        }
                        return object;
                    };
    
                    /**
                     * Converts this TypedValueList to JSON.
                     * @function toJSON
                     * @memberof fission.workflows.types.TypedValueList
                     * @instance
                     * @returns {Object.<string,*>} JSON object
                     */
                    TypedValueList.prototype.toJSON = function toJSON() {
                        return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                    };
    
                    return TypedValueList;
                })();
    
                return types;
            })();
    
            return workflows;
        })();
    
        return fission;
    })();
    
    $root.google = (function() {
    
        /**
         * Namespace google.
         * @exports google
         * @namespace
         */
        var google = {};
    
        google.protobuf = (function() {
    
            /**
             * Namespace protobuf.
             * @memberof google
             * @namespace
             */
            var protobuf = {};
    
            protobuf.Timestamp = (function() {
    
                /**
                 * Properties of a Timestamp.
                 * @memberof google.protobuf
                 * @interface ITimestamp
                 * @property {number|Long|null} [seconds] Timestamp seconds
                 * @property {number|null} [nanos] Timestamp nanos
                 */
    
                /**
                 * Constructs a new Timestamp.
                 * @memberof google.protobuf
                 * @classdesc Represents a Timestamp.
                 * @implements ITimestamp
                 * @constructor
                 * @param {google.protobuf.ITimestamp=} [properties] Properties to set
                 */
                function Timestamp(properties) {
                    if (properties)
                        for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                            if (properties[keys[i]] != null)
                                this[keys[i]] = properties[keys[i]];
                }
    
                /**
                 * Timestamp seconds.
                 * @member {number|Long} seconds
                 * @memberof google.protobuf.Timestamp
                 * @instance
                 */
                Timestamp.prototype.seconds = $util.Long ? $util.Long.fromBits(0,0,false) : 0;
    
                /**
                 * Timestamp nanos.
                 * @member {number} nanos
                 * @memberof google.protobuf.Timestamp
                 * @instance
                 */
                Timestamp.prototype.nanos = 0;
    
                /**
                 * Creates a new Timestamp instance using the specified properties.
                 * @function create
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {google.protobuf.ITimestamp=} [properties] Properties to set
                 * @returns {google.protobuf.Timestamp} Timestamp instance
                 */
                Timestamp.create = function create(properties) {
                    return new Timestamp(properties);
                };
    
                /**
                 * Encodes the specified Timestamp message. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
                 * @function encode
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {google.protobuf.ITimestamp} message Timestamp message or plain object to encode
                 * @param {$protobuf.Writer} [writer] Writer to encode to
                 * @returns {$protobuf.Writer} Writer
                 */
                Timestamp.encode = function encode(message, writer) {
                    if (!writer)
                        writer = $Writer.create();
                    if (message.seconds != null && message.hasOwnProperty("seconds"))
                        writer.uint32(/* id 1, wireType 0 =*/8).int64(message.seconds);
                    if (message.nanos != null && message.hasOwnProperty("nanos"))
                        writer.uint32(/* id 2, wireType 0 =*/16).int32(message.nanos);
                    return writer;
                };
    
                /**
                 * Encodes the specified Timestamp message, length delimited. Does not implicitly {@link google.protobuf.Timestamp.verify|verify} messages.
                 * @function encodeDelimited
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {google.protobuf.ITimestamp} message Timestamp message or plain object to encode
                 * @param {$protobuf.Writer} [writer] Writer to encode to
                 * @returns {$protobuf.Writer} Writer
                 */
                Timestamp.encodeDelimited = function encodeDelimited(message, writer) {
                    return this.encode(message, writer).ldelim();
                };
    
                /**
                 * Decodes a Timestamp message from the specified reader or buffer.
                 * @function decode
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                 * @param {number} [length] Message length if known beforehand
                 * @returns {google.protobuf.Timestamp} Timestamp
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                Timestamp.decode = function decode(reader, length) {
                    if (!(reader instanceof $Reader))
                        reader = $Reader.create(reader);
                    var end = length === undefined ? reader.len : reader.pos + length, message = new $root.google.protobuf.Timestamp();
                    while (reader.pos < end) {
                        var tag = reader.uint32();
                        switch (tag >>> 3) {
                        case 1:
                            message.seconds = reader.int64();
                            break;
                        case 2:
                            message.nanos = reader.int32();
                            break;
                        default:
                            reader.skipType(tag & 7);
                            break;
                        }
                    }
                    return message;
                };
    
                /**
                 * Decodes a Timestamp message from the specified reader or buffer, length delimited.
                 * @function decodeDelimited
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
                 * @returns {google.protobuf.Timestamp} Timestamp
                 * @throws {Error} If the payload is not a reader or valid buffer
                 * @throws {$protobuf.util.ProtocolError} If required fields are missing
                 */
                Timestamp.decodeDelimited = function decodeDelimited(reader) {
                    if (!(reader instanceof $Reader))
                        reader = new $Reader(reader);
                    return this.decode(reader, reader.uint32());
                };
    
                /**
                 * Verifies a Timestamp message.
                 * @function verify
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {Object.<string,*>} message Plain object to verify
                 * @returns {string|null} `null` if valid, otherwise the reason why it is not
                 */
                Timestamp.verify = function verify(message) {
                    if (typeof message !== "object" || message === null)
                        return "object expected";
                    if (message.seconds != null && message.hasOwnProperty("seconds"))
                        if (!$util.isInteger(message.seconds) && !(message.seconds && $util.isInteger(message.seconds.low) && $util.isInteger(message.seconds.high)))
                            return "seconds: integer|Long expected";
                    if (message.nanos != null && message.hasOwnProperty("nanos"))
                        if (!$util.isInteger(message.nanos))
                            return "nanos: integer expected";
                    return null;
                };
    
                /**
                 * Creates a Timestamp message from a plain object. Also converts values to their respective internal types.
                 * @function fromObject
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {Object.<string,*>} object Plain object
                 * @returns {google.protobuf.Timestamp} Timestamp
                 */
                Timestamp.fromObject = function fromObject(object) {
                    if (object instanceof $root.google.protobuf.Timestamp)
                        return object;
                    var message = new $root.google.protobuf.Timestamp();
                    if (object.seconds != null)
                        if ($util.Long)
                            (message.seconds = $util.Long.fromValue(object.seconds)).unsigned = false;
                        else if (typeof object.seconds === "string")
                            message.seconds = parseInt(object.seconds, 10);
                        else if (typeof object.seconds === "number")
                            message.seconds = object.seconds;
                        else if (typeof object.seconds === "object")
                            message.seconds = new $util.LongBits(object.seconds.low >>> 0, object.seconds.high >>> 0).toNumber();
                    if (object.nanos != null)
                        message.nanos = object.nanos | 0;
                    return message;
                };
    
                /**
                 * Creates a plain object from a Timestamp message. Also converts values to other types if specified.
                 * @function toObject
                 * @memberof google.protobuf.Timestamp
                 * @static
                 * @param {google.protobuf.Timestamp} message Timestamp
                 * @param {$protobuf.IConversionOptions} [options] Conversion options
                 * @returns {Object.<string,*>} Plain object
                 */
                Timestamp.toObject = function toObject(message, options) {
                    if (!options)
                        options = {};
                    var object = {};
                    if (options.defaults) {
                        if ($util.Long) {
                            var long = new $util.Long(0, 0, false);
                            object.seconds = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                        } else
                            object.seconds = options.longs === String ? "0" : 0;
                        object.nanos = 0;
                    }
                    if (message.seconds != null && message.hasOwnProperty("seconds"))
                        if (typeof message.seconds === "number")
                            object.seconds = options.longs === String ? String(message.seconds) : message.seconds;
                        else
                            object.seconds = options.longs === String ? $util.Long.prototype.toString.call(message.seconds) : options.longs === Number ? new $util.LongBits(message.seconds.low >>> 0, message.seconds.high >>> 0).toNumber() : message.seconds;
                    if (message.nanos != null && message.hasOwnProperty("nanos"))
                        object.nanos = message.nanos;
                    return object;
                };
    
                /**
                 * Converts this Timestamp to JSON.
                 * @function toJSON
                 * @memberof google.protobuf.Timestamp
                 * @instance
                 * @returns {Object.<string,*>} JSON object
                 */
                Timestamp.prototype.toJSON = function toJSON() {
                    return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
                };
    
                return Timestamp;
            })();
    
            return protobuf;
        })();
    
        return google;
    })();

    return $root;
});
