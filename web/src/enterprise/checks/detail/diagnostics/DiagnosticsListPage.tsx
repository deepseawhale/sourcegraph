import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import React from 'react'
import * as sourcegraph from 'sourcegraph'
import { Action } from '../../../../../../shared/src/api/types/action'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../../shared/src/util/errors'
import { withQueryParameter } from '../../../../components/withQueryParameter/WithQueryParameter'
import { ThemeProps } from '../../../../theme'
import { DiagnosticsListItem } from '../../../tasks/list/item/DiagnosticsListItem'
import { diagnosticID, DiagnosticInfo } from '../../../threads/detail/backend'
import { DiagnosticsBatchActionsButtonDropdown } from './changesets/DiagnosticsBatchActionsDropdownButton'
import { useDiagnostics } from './detail/useDiagnostics'
import { DiagnosticQueryBuilder } from './DiagnosticQueryBuilder'

interface Props extends ExtensionsControllerProps, PlatformContextProps, ThemeProps {
    baseDiagnosticQuery: sourcegraph.DiagnosticQuery
    selectedActions: { [diagnosticID: string]: Action | undefined }
    onActionSelect: (diagnostic: DiagnosticInfo, action: Action | null) => void

    className?: string
    history: H.History
    location: H.Location
}

const LOADING: 'loading' = 'loading'

/**
 * A page that lists diagnostics.
 */
export const DiagnosticsListPage = withQueryParameter<Props>(
    ({
        baseDiagnosticQuery,
        selectedActions,
        onActionSelect,
        query,
        onQueryChange,
        className = '',
        extensionsController,
        ...props
    }) => {
        const parsedQuery = parseDiagnosticQuery(baseDiagnosticQuery)
        // tslint:disable-next-line: react-hooks-nesting
        const diagnosticsOrError = useDiagnostics(extensionsController, parsedQuery)

        return (
            <div className={`diagnostics-list-page ${className}`}>
                <style>{`.diagnostics-list-page code, .diagnostics-list-page table, .diagnostics-list-page pre, .diagnostics-list-page .markdown, .diagnostics-list-page aside { line-height: 17.25px; background-color: black !important; margin: 0; padding: 0; border-spacing: 0; } .diagnostics-list-page pre { margin-top: 1px !important; margin-left: 9px !important;}`}</style>
                {isErrorLike(diagnosticsOrError) ? (
                    <div className="container">
                        <div className="alert alert-danger mt-2">{diagnosticsOrError.message}</div>
                    </div>
                ) : diagnosticsOrError === LOADING ? (
                    <div className="container">
                        <LoadingSpinner className="mt-3" />
                    </div>
                ) : diagnosticsOrError.length === 0 ? (
                    <div className="container">
                        <p className="p-2 mb-0 text-muted">No diagnostics found.</p>
                    </div>
                ) : (
                    <>
                        <DiagnosticQueryBuilder
                            {...props}
                            parsedQuery={parsedQuery}
                            extraActions={
                                <DiagnosticsBatchActionsButtonDropdown className="ml-5" buttonClassName="btn-primary" />
                            }
                            query={query}
                            onQueryChange={onQueryChange}
                        />
                        <ul className="list-group list-group-flush mb-0 border-top">
                            {diagnosticsOrError.map((diagnostic, i) => (
                                <li key={i} className={`list-group-item px-0 ${i === 0 ? 'border-top-0' : ''}`}>
                                    <DiagnosticsListItem
                                        {...props}
                                        key={JSON.stringify(diagnostic)}
                                        diagnostic={diagnostic}
                                        selectedAction={selectedActions[diagnosticID(diagnostic)] || null}
                                        onActionSelect={onActionSelect}
                                        className="container-fluid"
                                        extensionsController={extensionsController}
                                    />
                                </li>
                            ))}
                        </ul>
                    </>
                )}
            </div>
        )
    }
)

function parseDiagnosticQuery(base?: sourcegraph.DiagnosticQuery): sourcegraph.DiagnosticQuery {
    return base || { type: 'TODO!(sqs)' } // TODO!(sqs)
}
