/**
 * WorkerManager - Orchestrates all web workers for AgentFlow
 * Provides a unified interface for worker communication and load balancing
 */

class WorkerManager {
    constructor(baseUrl = '') {
        this.baseUrl = baseUrl;
        this.workers = {
            conversation: null,
            storage: null,
            message: null
        };
        this.isInitialized = false;
        this.messageId = 0;
        this.pendingMessages = new Map(); // Track pending worker messages
        this.workerLoadStats = {
            conversation: { active: 0, total: 0 },
            storage: { active: 0, total: 0 },
            message: { active: 0, total: 0 }
        };
    }

    /**
     * Initialize all workers
     */
    async init() {
        if (this.isInitialized) {
            return { success: true };
        }

        try {
            const workerPromises = [
                this.initWorker('conversation', `${this.baseUrl}/static/js/workers/conversationWorker.js`),
                this.initWorker('storage', `${this.baseUrl}/static/js/workers/storageWorker.js`),
                this.initWorker('message', `${this.baseUrl}/static/js/workers/messageWorker.js`)
            ];

            const results = await Promise.all(workerPromises);
            const failures = results.filter(r => !r.success);

            if (failures.length > 0) {
                throw new Error(`Failed to initialize workers: ${failures.map(f => f.error).join(', ')}`);
            }

            this.isInitialized = true;
            console.log('All workers initialized successfully');

            return { success: true };
        } catch (error) {
            console.error('Worker initialization failed:', error);
            return { success: false, error: error.message };
        }
    }

    /**
     * Initialize a single worker
     */
    async initWorker(type, scriptUrl) {
        return new Promise((resolve) => {
            try {
                console.log(`Attempting to create worker ${type} from ${scriptUrl}`);
                const worker = new Worker(scriptUrl);
                let isReady = false;

                worker.onmessage = (e) => {
                    console.log(`Worker ${type} message received:`, e.data);
                    const { type: messageType, id } = e.data;

                    if (messageType === 'ready' && !isReady) {
                        console.log(`Worker ${type} is ready!`);
                        isReady = true;
                        this.workers[type] = worker;
                        this.setupWorkerEventHandling(worker, type);
                        resolve({ success: true });
                    } else if (id && this.pendingMessages.has(id)) {
                        // Handle regular worker responses
                        const { resolve: msgResolve } = this.pendingMessages.get(id);
                        this.pendingMessages.delete(id);
                        this.workerLoadStats[type].active--;
                        msgResolve(e.data);
                    }
                };

                worker.onerror = (error) => {
                    console.warn(`Worker ${type} error event (may be recoverable):`, error);
                    // Don't fail immediately on error - wait for ready message or timeout
                    if (isReady) {
                        console.error(`Worker ${type} runtime error:`, error);
                    }
                };

                // Timeout for worker initialization (increased to 10 seconds)
                setTimeout(() => {
                    if (!isReady) {
                        console.error(`Worker ${type} timeout - terminating`);
                        worker.terminate();
                        resolve({ success: false, error: `Worker ${type} initialization timeout` });
                    }
                }, 10000);

            } catch (error) {
                resolve({ success: false, error: `Failed to create worker ${type}: ${error.message}` });
            }
        });
    }

    /**
     * Setup additional event handling for worker
     */
    setupWorkerEventHandling(worker, type) {
        worker.addEventListener('error', (error) => {
            console.error(`Worker ${type} error:`, error);
            this.handleWorkerError(type, error);
        });

        worker.addEventListener('messageerror', (error) => {
            console.error(`Worker ${type} message error:`, error);
        });
    }

    /**
     * Handle worker errors and recovery
     */
    handleWorkerError(type, error) {
        console.warn(`Worker ${type} encountered an error, attempting recovery...`);

        // Clear pending messages for this worker
        for (const [id, pending] of this.pendingMessages.entries()) {
            if (pending.workerType === type) {
                pending.reject(new Error(`Worker ${type} failed: ${error.message}`));
                this.pendingMessages.delete(id);
            }
        }

        // Reset load stats
        this.workerLoadStats[type].active = 0;

        // Attempt to restart worker
        setTimeout(() => {
            this.restartWorker(type);
        }, 1000);
    }

    /**
     * Restart a failed worker
     */
    async restartWorker(type) {
        try {
            if (this.workers[type]) {
                this.workers[type].terminate();
                this.workers[type] = null;
            }

            const scriptUrls = {
                conversation: `${this.baseUrl}/static/js/workers/conversationWorker.js`,
                storage: `${this.baseUrl}/static/js/workers/storageWorker.js`,
                message: `${this.baseUrl}/static/js/workers/messageWorker.js`
            };

            const result = await this.initWorker(type, scriptUrls[type]);
            if (result.success) {
                console.log(`Worker ${type} restarted successfully`);
            } else {
                console.error(`Failed to restart worker ${type}:`, result.error);
            }

            return result;
        } catch (error) {
            console.error(`Error restarting worker ${type}:`, error);
            return { success: false, error: error.message };
        }
    }

    /**
     * Send message to worker with promise-based response
     */
    async sendToWorker(workerType, messageType, data, timeout = 30000) {
        if (!this.isInitialized) {
            throw new Error('Workers not initialized');
        }

        const worker = this.workers[workerType];
        if (!worker) {
            throw new Error(`Worker ${workerType} not available`);
        }

        return new Promise((resolve, reject) => {
            const id = ++this.messageId;

            // Store the pending message
            this.pendingMessages.set(id, {
                resolve,
                reject,
                workerType,
                timestamp: Date.now()
            });

            // Update load stats
            this.workerLoadStats[workerType].active++;
            this.workerLoadStats[workerType].total++;

            // Send message to worker
            worker.postMessage({
                id,
                type: messageType,
                data
            });

            // Set timeout
            setTimeout(() => {
                if (this.pendingMessages.has(id)) {
                    this.pendingMessages.delete(id);
                    this.workerLoadStats[workerType].active--;
                    reject(new Error(`Worker ${workerType} timeout for ${messageType}`));
                }
            }, timeout);
        });
    }

    /**
     * Conversation worker methods
     */
    async processConversationForStorage(conversationData) {
        return this.sendToWorker('conversation', 'processForStorage', conversationData);
    }

    async prepareConversationForAPI(conversation, systemPrompt, selectedTools) {
        return this.sendToWorker('conversation', 'prepareForAPI', {
            conversation,
            systemPrompt,
            selectedTools
        });
    }

    async calculateConversationStats(conversations) {
        return this.sendToWorker('conversation', 'calculateStats', conversations);
    }

    async suggestConversationCleanup(conversations, maxSize) {
        return this.sendToWorker('conversation', 'suggestCleanup', {
            conversations,
            maxSize
        });
    }

    /**
     * Storage worker methods (data processing only - not actual storage)
     */
    async processConversationsForStorage(conversations) {
        return this.sendToWorker('storage', 'processConversationsForStorage', conversations);
    }

    async createReducedConversations(conversations) {
        return this.sendToWorker('storage', 'createReducedConversations', conversations);
    }

    async createEmergencyData(conversations) {
        return this.sendToWorker('storage', 'createEmergencyData', conversations);
    }

    async suggestCleanup(conversations, maxCount, maxAge) {
        return this.sendToWorker('storage', 'suggestCleanup', {
            conversations,
            maxCount,
            maxAge
        });
    }

    async calculateStorageUsage(conversations) {
        return this.sendToWorker('storage', 'calculateStorageUsage', conversations);
    }

    /**
     * Message worker methods
     */
    async processMessageContent(message) {
        return this.sendToWorker('message', 'processMessageContent', message);
    }

    async processMessagesForSearch(messages, searchTerm) {
        return this.sendToWorker('message', 'processMessagesForSearch', {
            messages,
            searchTerm
        });
    }

    async processConversationForExport(conversation, format) {
        return this.sendToWorker('message', 'processConversationForExport', {
            conversation,
            format
        });
    }

    async processFileAttachment(file) {
        return this.sendToWorker('message', 'processFileAttachment', file);
    }

    /**
     * Get worker performance stats
     */
    getPerformanceStats() {
        return {
            workers: Object.keys(this.workers).map(type => ({
                type,
                available: !!this.workers[type],
                activeJobs: this.workerLoadStats[type].active,
                totalJobsProcessed: this.workerLoadStats[type].total
            })),
            pendingMessages: this.pendingMessages.size,
            initialized: this.isInitialized
        };
    }

    /**
     * Cleanup and terminate all workers
     */
    terminate() {
        Object.values(this.workers).forEach(worker => {
            if (worker) {
                worker.terminate();
            }
        });

        this.workers = {
            conversation: null,
            storage: null,
            message: null
        };

        this.isInitialized = false;
        this.pendingMessages.clear();

        console.log('All workers terminated');
    }

    /**
     * Health check for all workers
     */
    async healthCheck() {
        if (!this.isInitialized) {
            return { healthy: false, error: 'Not initialized' };
        }

        try {
            const checks = await Promise.allSettled([
                this.sendToWorker('conversation', 'calculateStats', {}),
                this.sendToWorker('storage', 'calculateStorageUsage', {}),
                this.sendToWorker('message', 'processMessageContent', { content: 'test' })
            ]);

            const failures = checks.filter(check => check.status === 'rejected');

            return {
                healthy: failures.length === 0,
                results: checks,
                stats: this.getPerformanceStats()
            };
        } catch (error) {
            return {
                healthy: false,
                error: error.message,
                stats: this.getPerformanceStats()
            };
        }
    }
}

// Export for use in main chat application
window.WorkerManager = WorkerManager;