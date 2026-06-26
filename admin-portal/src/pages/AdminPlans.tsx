import { useEffect, useState, type FormEvent } from 'react';
import { api } from '../lib/api';
import type { AudienceSegment, BillingRate, CreatePlanRequest, CreateSegmentRequest } from '../types';
import { Plus, DollarSign, Layers } from 'lucide-react';

type Tab = 'plans' | 'segments';

interface Plan {
  id: string;
  name: string;
  monthly_fee_etb: number;
  max_active_campaigns: number;
  max_daily_budget_etb: number;
  included_impressions: number;
  sms_plus_enabled: boolean;
  audiencemart_enabled: boolean;
}

export default function AdminPlans() {
  const [tab, setTab] = useState<Tab>('plans');
  const [segments, setSegments] = useState<AudienceSegment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  // Plan form state
  const [showPlanForm, setShowPlanForm] = useState(false);
  const [planLoading, setPlanLoading] = useState(false);
  const [planForm, setPlanForm] = useState<CreatePlanRequest>({
    name: '', monthly_fee_etb: 0, max_active_campaigns: 5,
    max_daily_budget_etb: 100, included_impressions: 10000,
    sms_plus_enabled: false, audiencemart_enabled: false, cpc_discount_pct: 0,
  });

  // Segment form state
  const [showSegForm, setShowSegForm] = useState(false);
  const [segLoading, setSegLoading] = useState(false);
  const [segForm, setSegForm] = useState<CreateSegmentRequest>({
    name: '', description: '', top_intent_signals: [],
    approximate_size: 0, estimated_cpm: 0, is_active: true,
  });
  const [signalsInput, setSignalsInput] = useState('');

  // Billing rates drill-down
  const [selectedPlanId, setSelectedPlanId] = useState<string | null>(null);
  const [rates, setRates] = useState<BillingRate[]>([]);
  const [ratesLoading, setRatesLoading] = useState(false);

  // Plans list (from /plans public endpoint)
  const [plans, setPlans] = useState<Plan[]>([]);

  useEffect(() => {
    (async () => {
      try {
        setLoading(true);
        const [planList, segList] = await Promise.all([
          api.listCampaigns(0, 0).then(() => [] as Plan[]).catch(() => [] as Plan[]),
          api.listSegments(),
        ]);
        // Try loading plans from the public endpoint
        try {
          const res = await fetch('/api/v1/ad-portal/plans', {
            headers: { Authorization: `Bearer ${localStorage.getItem('adminPortalToken') || ''}` },
          });
          if (res.ok) {
            const body = await res.json();
            setPlans(body.plans ?? []);
          }
        } catch { setPlans(planList); }
        setSegments(segList);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load data');
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  async function handleCreatePlan(e: FormEvent) {
    e.preventDefault();
    setPlanLoading(true);
    setMessage(null);
    try {
      await api.createPlan(planForm);
      setMessage({ type: 'success', text: `Plan "${planForm.name}" created successfully.` });
      setShowPlanForm(false);
      setPlanForm({ name: '', monthly_fee_etb: 0, max_active_campaigns: 5, max_daily_budget_etb: 100, included_impressions: 10000, sms_plus_enabled: false, audiencemart_enabled: false, cpc_discount_pct: 0 });
      // Refresh plans
      try {
        const res = await fetch('/api/v1/ad-portal/plans', {
          headers: { Authorization: `Bearer ${localStorage.getItem('adminPortalToken') || ''}` },
        });
        if (res.ok) { const b = await res.json(); setPlans(b.plans ?? []); }
      } catch {}
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Failed to create plan' });
    } finally {
      setPlanLoading(false);
    }
  }

  async function handleCreateSegment(e: FormEvent) {
    e.preventDefault();
    setSegLoading(true);
    setMessage(null);
    const data = { ...segForm, top_intent_signals: signalsInput.split(',').map(s => s.trim()).filter(Boolean) };
    try {
      await api.createSegment(data);
      setMessage({ type: 'success', text: `Segment "${segForm.name}" created.` });
      setShowSegForm(false);
      setSegForm({ name: '', description: '', top_intent_signals: [], approximate_size: 0, estimated_cpm: 0, is_active: true });
      setSignalsInput('');
      const segList = await api.listSegments();
      setSegments(segList);
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Failed to create segment' });
    } finally {
      setSegLoading(false);
    }
  }

  async function loadBillingRates(planId: string) {
    setSelectedPlanId(planId);
    setRatesLoading(true);
    try {
      const res = await api.listBillingRates(planId);
      setRates(res.rates ?? []);
    } catch (err) {
      setMessage({ type: 'error', text: err instanceof Error ? err.message : 'Failed to load rates' });
    } finally {
      setRatesLoading(false);
    }
  }

  if (loading) return <div className="text-muted text-sm">Loading…</div>;

  return (
    <div>
      <h1 className="text-xl font-bold text-primary mb-1">Plans & Catalog</h1>
      <p className="text-sm text-muted mb-5">Manage subscription plans, billing rates, and audience segments.</p>

      {message && (
        <div className={`mb-4 p-3 rounded-lg border text-xs ${
          message.type === 'success' ? 'bg-green-50 border-green-200 text-green-700 dark:bg-green-900/20 dark:border-green-800 dark:text-green-400'
            : 'bg-red-50 border-red-200 text-red-700 dark:bg-red-900/20 dark:border-red-800 dark:text-red-400'
        }`}>{message.text}</div>
      )}
      {error && <div className="alert-error mb-4 text-sm">{error}</div>}

      <div className="tab-bar mb-5">
        <button type="button" className={`tab-btn text-xs ${tab === 'plans' ? 'tab-btn-active' : ''}`} onClick={() => setTab('plans')}>
          <DollarSign size={13} /> Plans & Billing
        </button>
        <button type="button" className={`tab-btn text-xs ${tab === 'segments' ? 'tab-btn-active' : ''}`} onClick={() => setTab('segments')}>
          <Layers size={13} /> Audience Segments
        </button>
      </div>

      {/* Plans Tab */}
      {tab === 'plans' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-primary">Subscription Plans</h2>
            <button type="button" onClick={() => setShowPlanForm(!showPlanForm)} className="btn-primary text-xs">
              <Plus size={13} /> New Plan
            </button>
          </div>

          {showPlanForm && (
            <form onSubmit={handleCreatePlan} className="card-static p-5 mb-5 space-y-4">
              <div className="grid sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">Plan Name</label>
                  <input required value={planForm.name} onChange={e => setPlanForm({ ...planForm, name: e.target.value })} className="field-input text-sm" placeholder="Pro" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">Monthly Fee (ETB)</label>
                  <input type="number" step="0.01" required value={planForm.monthly_fee_etb || ''} onChange={e => setPlanForm({ ...planForm, monthly_fee_etb: parseFloat(e.target.value) || 0 })} className="field-input text-sm" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">Max Active Campaigns</label>
                  <input type="number" required value={planForm.max_active_campaigns} onChange={e => setPlanForm({ ...planForm, max_active_campaigns: parseInt(e.target.value) || 1 })} className="field-input text-sm" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">Max Daily Budget (ETB)</label>
                  <input type="number" step="0.01" required value={planForm.max_daily_budget_etb || ''} onChange={e => setPlanForm({ ...planForm, max_daily_budget_etb: parseFloat(e.target.value) || 0 })} className="field-input text-sm" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">Included Impressions</label>
                  <input type="number" value={planForm.included_impressions} onChange={e => setPlanForm({ ...planForm, included_impressions: parseInt(e.target.value) || 0 })} className="field-input text-sm" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">CPC Discount %</label>
                  <input type="number" step="0.1" value={planForm.cpc_discount_pct || ''} onChange={e => setPlanForm({ ...planForm, cpc_discount_pct: parseFloat(e.target.value) || 0 })} className="field-input text-sm" />
                </div>
              </div>
              <div className="flex gap-4">
                <label className="flex items-center gap-2 text-xs text-primary cursor-pointer">
                  <input type="checkbox" checked={planForm.sms_plus_enabled} onChange={e => setPlanForm({ ...planForm, sms_plus_enabled: e.target.checked })} className="rounded" />
                  SMS+ Enabled
                </label>
                <label className="flex items-center gap-2 text-xs text-primary cursor-pointer">
                  <input type="checkbox" checked={planForm.audiencemart_enabled} onChange={e => setPlanForm({ ...planForm, audiencemart_enabled: e.target.checked })} className="rounded" />
                  AudienceMart Enabled
                </label>
              </div>
              <div className="flex gap-2 pt-2 border-t border-[var(--border)]">
                <button type="submit" disabled={planLoading} className="btn-primary text-xs">{planLoading ? 'Creating…' : 'Create Plan'}</button>
                <button type="button" onClick={() => setShowPlanForm(false)} className="btn-secondary text-xs">Cancel</button>
              </div>
            </form>
          )}

          {plans.length === 0 ? (
            <div className="card-static p-8 text-center border-dashed">
              <p className="text-sm text-muted">No plans found. Create the first subscription plan.</p>
            </div>
          ) : (
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
              {plans.map(p => (
                <div key={p.id} className="card-static p-4">
                  <h3 className="text-sm font-semibold text-primary">{p.name}</h3>
                  <p className="text-lg font-bold text-brand-600 dark:text-brand-400 mt-1">ETB {p.monthly_fee_etb}/mo</p>
                  <div className="text-xs text-muted mt-2 space-y-0.5">
                    <p>Max {p.max_active_campaigns} active campaigns</p>
                    <p>Daily budget cap: ETB {p.max_daily_budget_etb}</p>
                    <p>{p.included_impressions.toLocaleString()} impressions</p>
                  </div>
                  <div className="flex gap-2 mt-3 flex-wrap">
                    {p.sms_plus_enabled && <span className="text-[10px] px-1.5 py-0.5 rounded bg-blue-50 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400 border border-blue-200 dark:border-blue-800">SMS+</span>}
                    {p.audiencemart_enabled && <span className="text-[10px] px-1.5 py-0.5 rounded bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400 border border-green-200 dark:border-green-800">AudienceMart</span>}
                  </div>
                  <button type="button" onClick={() => loadBillingRates(p.id)} className="btn-ghost text-xs mt-3 w-full">
                    View Billing Rates
                  </button>
                </div>
              ))}
            </div>
          )}

          {/* Billing Rates Drill-Down */}
          {selectedPlanId && (
            <div className="mt-5 card-static p-5">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-primary">Billing Rates</h3>
                <button type="button" onClick={() => setSelectedPlanId(null)} className="text-xs text-muted hover:text-primary">Close</button>
              </div>
              {ratesLoading ? (
                <p className="text-sm text-muted">Loading rates…</p>
              ) : rates.length === 0 ? (
                <p className="text-sm text-muted">No billing rates configured for this plan.</p>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full text-left border-collapse text-sm">
                    <thead>
                      <tr className="border-b border-[var(--border)] text-[11px] uppercase tracking-wider text-muted">
                        <th className="p-2.5 font-medium">Model</th>
                        <th className="p-2.5 font-medium text-right">Rate (ETB)</th>
                        <th className="p-2.5 font-medium text-center">Status</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-[var(--border)]">
                      {rates.map(r => (
                        <tr key={r.id} className="hover:bg-[var(--bg-subtle)]">
                          <td className="p-2.5 text-primary text-xs font-medium">{r.billing_model}</td>
                          <td className="p-2.5 text-right text-primary text-xs">{r.rate_etb.toFixed(2)}</td>
                          <td className="p-2.5 text-center">
                            <span className={`inline-flex px-1.5 py-0.5 rounded text-[10px] font-medium ${r.is_active ? 'bg-green-50 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-gray-100 text-gray-500 dark:bg-gray-800 dark:text-gray-400'}`}>
                              {r.is_active ? 'Active' : 'Inactive'}
                            </span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          )}
        </div>
      )}

      {/* Segments Tab */}
      {tab === 'segments' && (
        <div>
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-primary">Audience Segments ({segments.length})</h2>
            <button type="button" onClick={() => setShowSegForm(!showSegForm)} className="btn-primary text-xs">
              <Plus size={13} /> New Segment
            </button>
          </div>

          {showSegForm && (
            <form onSubmit={handleCreateSegment} className="card-static p-5 mb-5 space-y-4">
              <div className="grid sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">Segment Name</label>
                  <input required value={segForm.name} onChange={e => setSegForm({ ...segForm, name: e.target.value })} className="field-input text-sm" placeholder="Coffee Lovers" />
                </div>
                <div>
                  <label className="block text-xs font-medium text-primary mb-1">Estimated CPM (ETB)</label>
                  <input type="number" step="0.01" required value={segForm.estimated_cpm || ''} onChange={e => setSegForm({ ...segForm, estimated_cpm: parseFloat(e.target.value) || 0 })} className="field-input text-sm" />
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-primary mb-1">Description</label>
                <textarea value={segForm.description} onChange={e => setSegForm({ ...segForm, description: e.target.value })} rows={2} className="field-input text-sm resize-y" />
              </div>
              <div>
                <label className="block text-xs font-medium text-primary mb-1">Intent Signals (comma separated)</label>
                <input value={signalsInput} onChange={e => setSignalsInput(e.target.value)} className="field-input text-sm" placeholder="crypto_interest, fintech_interest" />
              </div>
              <div className="flex gap-2 pt-2 border-t border-[var(--border)]">
                <button type="submit" disabled={segLoading} className="btn-primary text-xs">{segLoading ? 'Creating…' : 'Create Segment'}</button>
                <button type="button" onClick={() => setShowSegForm(false)} className="btn-secondary text-xs">Cancel</button>
              </div>
            </form>
          )}

          {segments.length === 0 ? (
            <div className="card-static p-8 text-center border-dashed">
              <p className="text-sm text-muted">No audience segments found.</p>
            </div>
          ) : (
            <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-4">
              {segments.map(seg => (
                <div key={seg.id} className="card-static p-4">
                  <h3 className="text-sm font-semibold text-primary">{seg.name}</h3>
                  <p className="text-[10px] font-mono text-muted mt-0.5 break-all">{seg.id}</p>
                  {seg.description && <p className="text-xs text-muted mt-1.5">{seg.description}</p>}
                  <div className="mt-3 pt-3 border-t border-[var(--border)] flex justify-between items-center">
                    <span className="text-xs font-medium text-muted">Price (ETB)</span>
                    <span className="text-sm font-bold text-brand-600 dark:text-brand-400">{seg.price_etb}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
