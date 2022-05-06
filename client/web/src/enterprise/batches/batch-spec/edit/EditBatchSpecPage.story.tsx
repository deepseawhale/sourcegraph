import { storiesOf } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import {
    EMPTY_SETTINGS_CASCADE,
    SettingsOrgSubject,
    SettingsUserSubject,
} from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../../components/WebStory'
import { GET_BATCH_CHANGE_TO_EDIT } from '../../create/backend'
import { insertNameIntoLibraryItem } from '../../create/yaml-util'
import { mockBatchChange, mockBatchSpec } from '../batch-spec.mock'

import { EditBatchSpecPage } from './EditBatchSpecPage'
import goImportsSample from './library/go-imports.batch.yaml'

const { add } = storiesOf('web/batches/batch-spec/edit/EditBatchSpecPage', module)
    .addDecorator(story => <div className="p-3 w-100">{story()}</div>)
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

const FIXTURE_ORG: SettingsOrgSubject = {
    __typename: 'Org',
    name: 'sourcegraph',
    displayName: 'Sourcegraph',
    id: 'a',
    viewerCanAdminister: true,
}

const FIXTURE_USER: SettingsUserSubject = {
    __typename: 'User',
    username: 'alice',
    displayName: 'alice',
    id: 'b',
    viewerCanAdminister: true,
}

const SETTINGS_CASCADE = {
    ...EMPTY_SETTINGS_CASCADE,
    subjects: [
        { subject: FIXTURE_ORG, settings: { a: 1 }, lastID: 1 },
        { subject: FIXTURE_USER, settings: { b: 2 }, lastID: 2 },
    ],
}

const FIRST_TIME_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: { batchChange: mockBatchChange() } },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

add('editing for the first time', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={FIRST_TIME_MOCKS}>
                <div style={{ height: '95vh', width: '100%' }}>
                    <EditBatchSpecPage
                        {...props}
                        batchChange={{
                            name: 'my-batch-change',
                            url: 'some-url',
                            namespace: { id: 'test1234' },
                        }}
                        settingsCascade={SETTINGS_CASCADE}
                    />
                </div>
            </MockedTestProvider>
        )}
    </WebStory>
))

const MULTIPLE_SPEC_MOCKS = new WildcardMockLink([
    {
        request: {
            query: getDocumentNode(GET_BATCH_CHANGE_TO_EDIT),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: {
            data: {
                batchChange: mockBatchChange({
                    batchSpecs: {
                        nodes: [
                            mockBatchSpec({
                                id: 'new',
                                originalInput: insertNameIntoLibraryItem(goImportsSample, 'my-batch-change'),
                            }),
                            mockBatchSpec({ id: 'old1' }),
                            mockBatchSpec({ id: 'old2' }),
                        ],
                    },
                }),
            },
        },
        nMatches: Number.POSITIVE_INFINITY,
    },
])

add('editing the latest batch spec', () => (
    <WebStory>
        {props => (
            <MockedTestProvider link={MULTIPLE_SPEC_MOCKS}>
                <div style={{ height: '95vh', width: '100%' }}>
                    <EditBatchSpecPage
                        {...props}
                        batchChange={{
                            name: 'my-batch-change',
                            url: 'some-url',
                            namespace: { id: 'test1234' },
                        }}
                        settingsCascade={SETTINGS_CASCADE}
                    />
                </div>
            </MockedTestProvider>
        )}
    </WebStory>
))

add('batch change not found', () => (
    <WebStory>
        {props => (
            <div style={{ height: '95vh', width: '100%' }}>
                <EditBatchSpecPage
                    {...props}
                    batchChange={{
                        name: 'doesnt-exist',
                        url: 'some-url',
                        namespace: { id: 'test1234' },
                    }}
                    settingsCascade={SETTINGS_CASCADE}
                />
            </div>
        )}
    </WebStory>
))
